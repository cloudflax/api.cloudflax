package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	shareddynamodb "github.com/cloudflax/api.cloudflax/internal/shared/dynamodb"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	resendCooldownSeconds = int64(5 * 60)
	resendMaxSends        = int64(3)
	resendLockSeconds     = int64(2 * 60 * 60)
	resendStateTTLSeconds = int64(24 * 60 * 60)
	resendStateSK         = "STATE"
)

// En: ErrResendVerificationRateLimited indicates resend throttle limits were reached.
// Es: ErrResendVerificationRateLimited indica que se alcanzaron limites de reenvio.
var ErrResendVerificationRateLimited = errors.New("resend verification rate limited")

// En: ResendVerificationGuard enforces resend verification throttling policies.
// Es: ResendVerificationGuard aplica politicas de throttling para reenvio.
type ResendVerificationGuard interface {
	CheckAndConsume(ctx context.Context, email, ip string) error
}

// En: DynamoResendVerificationGuardOptions configures the DynamoDB resend guard.
// Es: DynamoResendVerificationGuardOptions configura el guard de reenvio en DynamoDB.
type DynamoResendVerificationGuardOptions struct {
	TableName       string
	EndpointURL     string
	Region          string
	Profile         string
	AccessKeyID     string
	SecretAccessKey string
}

// En: ResendVerificationRateLimitError includes retry delay for throttled resend calls.
// Es: ResendVerificationRateLimitError incluye demora de reintento para reenvio bloqueado.
type ResendVerificationRateLimitError struct {
	RetryAfter time.Duration
}

// En: Error returns the throttle reason message.
// Es: Error devuelve el mensaje del motivo del throttle.
func (e *ResendVerificationRateLimitError) Error() string {
	return ErrResendVerificationRateLimited.Error()
}

// En: Is reports whether err is a resend verification rate-limit error.
// Es: Is indica si err es un error de limite de reenvio.
func (e *ResendVerificationRateLimitError) Is(target error) bool {
	return target == ErrResendVerificationRateLimited
}

type resendGuardState struct {
	Exists         bool
	Version        int64
	Count          int64
	WindowStart    int64
	NextAllowedAt  int64
	LockUntil      int64
	CreatedAt      int64
	UpdatedAt      int64
	ExpiresAt      int64
	OriginalPK     string
	OriginalSK     string
	OriginalItem   map[string]types.AttributeValue
	HasLockUntil   bool
	HasCount       bool
	HasVersion     bool
	HasWindowStart bool
}

type dynamoAPI interface {
	GetItem(ctx context.Context, params *awsdynamodb.GetItemInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *awsdynamodb.PutItemInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error)
}

// En: dynamoResendVerificationGuard stores throttle state in DynamoDB.
// Es: dynamoResendVerificationGuard guarda estado de throttle en DynamoDB.
type dynamoResendVerificationGuard struct {
	client    dynamoAPI
	tableName string
	now       func() time.Time
}

// En: NewDynamoResendVerificationGuard builds a DynamoDB-backed resend guard.
// Es: NewDynamoResendVerificationGuard crea un guard de reenvio con DynamoDB.
func NewDynamoResendVerificationGuard(ctx context.Context, opts DynamoResendVerificationGuardOptions) (ResendVerificationGuard, error) {
	tableName := strings.TrimSpace(opts.TableName)
	if tableName == "" {
		return nil, nil
	}

	client, err := shareddynamodb.NewClient(ctx, shareddynamodb.ClientOptions{
		EndpointURL:     opts.EndpointURL,
		Region:          opts.Region,
		Profile:         opts.Profile,
		AccessKeyID:     opts.AccessKeyID,
		SecretAccessKey: opts.SecretAccessKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create dynamodb client for resend guard: %w", err)
	}

	return &dynamoResendVerificationGuard{
		client:    client,
		tableName: tableName,
		now:       time.Now,
	}, nil
}

// En: CheckAndConsume validates limits and consumes one resend quota atomically.
// Es: CheckAndConsume valida limites y consume una cuota de reenvio.
func (g *dynamoResendVerificationGuard) CheckAndConsume(ctx context.Context, email, _ string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" {
		return nil
	}

	emailHash := sha256Hex(normalizedEmail)
	pk := fmt.Sprintf("THROTTLE#RESEND_VERIFICATION#EMAIL#%s", emailHash)

	const maxRetries = 4
	for range maxRetries {
		now := g.now().Unix()
		current, err := g.getState(ctx, pk)
		if err != nil {
			return fmt.Errorf("load resend throttle state: %w", err)
		}

		next, limitErr := evaluateResendState(current, now)
		if limitErr != nil {
			return limitErr
		}
		if next.OriginalPK == "" {
			next.OriginalPK = pk
		}
		if next.OriginalSK == "" {
			next.OriginalSK = resendStateSK
		}

		if err := g.putState(ctx, current, next); err != nil {
			var ccf *types.ConditionalCheckFailedException
			if errors.As(err, &ccf) {
				continue
			}
			return fmt.Errorf("store resend throttle state: %w", err)
		}
		return nil
	}

	return fmt.Errorf("store resend throttle state: too much contention")
}

func (g *dynamoResendVerificationGuard) getState(ctx context.Context, pk string) (*resendGuardState, error) {
	out, err := g.client.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName:      aws.String(g.tableName),
		ConsistentRead: aws.Bool(true),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
			"sk": &types.AttributeValueMemberS{Value: resendStateSK},
		},
	})
	if err != nil {
		return nil, err
	}

	state := &resendGuardState{}
	if len(out.Item) == 0 {
		return state, nil
	}

	state.Exists = true
	state.OriginalItem = out.Item
	state.OriginalPK = attrString(out.Item["pk"])
	state.OriginalSK = attrString(out.Item["sk"])
	state.Version, state.HasVersion = attrInt64(out.Item["version"])
	state.Count, state.HasCount = attrInt64(out.Item["count"])
	state.WindowStart, state.HasWindowStart = attrInt64(out.Item["window_start"])
	state.NextAllowedAt, _ = attrInt64(out.Item["next_allowed_at"])
	state.LockUntil, state.HasLockUntil = attrInt64(out.Item["lock_until"])
	state.CreatedAt, _ = attrInt64(out.Item["created_at"])
	state.UpdatedAt, _ = attrInt64(out.Item["updated_at"])
	state.ExpiresAt, _ = attrInt64(out.Item["expires_at"])

	return state, nil
}

func (g *dynamoResendVerificationGuard) putState(ctx context.Context, current, next *resendGuardState) error {
	item := map[string]types.AttributeValue{
		"pk":              &types.AttributeValueMemberS{Value: next.OriginalPK},
		"sk":              &types.AttributeValueMemberS{Value: next.OriginalSK},
		"version":         &types.AttributeValueMemberN{Value: strconv.FormatInt(next.Version, 10)},
		"count":           &types.AttributeValueMemberN{Value: strconv.FormatInt(next.Count, 10)},
		"window_start":    &types.AttributeValueMemberN{Value: strconv.FormatInt(next.WindowStart, 10)},
		"next_allowed_at": &types.AttributeValueMemberN{Value: strconv.FormatInt(next.NextAllowedAt, 10)},
		"created_at":      &types.AttributeValueMemberN{Value: strconv.FormatInt(next.CreatedAt, 10)},
		"updated_at":      &types.AttributeValueMemberN{Value: strconv.FormatInt(next.UpdatedAt, 10)},
		"expires_at":      &types.AttributeValueMemberN{Value: strconv.FormatInt(next.ExpiresAt, 10)},
	}
	if next.LockUntil > 0 {
		item["lock_until"] = &types.AttributeValueMemberN{Value: strconv.FormatInt(next.LockUntil, 10)}
	}

	input := &awsdynamodb.PutItemInput{
		TableName: aws.String(g.tableName),
		Item:      item,
	}
	if current.Exists {
		input.ConditionExpression = aws.String("#version = :expectedVersion")
		input.ExpressionAttributeNames = map[string]string{
			"#version": "version",
		}
		input.ExpressionAttributeValues = map[string]types.AttributeValue{
			":expectedVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(current.Version, 10)},
		}
	} else {
		input.ConditionExpression = aws.String("attribute_not_exists(pk) AND attribute_not_exists(sk)")
	}

	_, err := g.client.PutItem(ctx, input)
	return err
}

func evaluateResendState(current *resendGuardState, now int64) (*resendGuardState, error) {
	if current == nil || !current.Exists {
		return buildFreshState(now), nil
	}

	if current.LockUntil > now {
		return nil, &ResendVerificationRateLimitError{
			RetryAfter: time.Duration(current.LockUntil-now) * time.Second,
		}
	}

	if current.Count >= resendMaxSends {
		return buildFreshStateFrom(current, now), nil
	}

	if current.NextAllowedAt > now {
		return nil, &ResendVerificationRateLimitError{
			RetryAfter: time.Duration(current.NextAllowedAt-now) * time.Second,
		}
	}

	next := buildFreshStateFrom(current, now)
	next.Count = current.Count + 1
	next.NextAllowedAt = now + resendCooldownSeconds
	next.ExpiresAt = next.NextAllowedAt + resendStateTTLSeconds
	if next.Count >= resendMaxSends {
		next.LockUntil = now + resendLockSeconds
		next.ExpiresAt = next.LockUntil + resendStateTTLSeconds
	}

	return next, nil
}

func buildFreshState(now int64) *resendGuardState {
	return &resendGuardState{
		OriginalPK:    "",
		OriginalSK:    resendStateSK,
		Version:       1,
		Count:         1,
		WindowStart:   now,
		NextAllowedAt: now + resendCooldownSeconds,
		CreatedAt:     now,
		UpdatedAt:     now,
		ExpiresAt:     now + resendCooldownSeconds + resendStateTTLSeconds,
	}
}

func buildFreshStateFrom(current *resendGuardState, now int64) *resendGuardState {
	if current == nil {
		return buildFreshState(now)
	}

	createdAt := current.CreatedAt
	if createdAt == 0 {
		createdAt = now
	}

	return &resendGuardState{
		Exists:        current.Exists,
		OriginalPK:    current.OriginalPK,
		OriginalSK:    current.OriginalSK,
		Version:       current.Version + 1,
		Count:         1,
		WindowStart:   now,
		CreatedAt:     createdAt,
		UpdatedAt:     now,
		NextAllowedAt: now + resendCooldownSeconds,
		ExpiresAt:     now + resendCooldownSeconds + resendStateTTLSeconds,
	}
}

func attrInt64(v types.AttributeValue) (int64, bool) {
	if v == nil {
		return 0, false
	}
	nv, ok := v.(*types.AttributeValueMemberN)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(nv.Value, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func attrString(v types.AttributeValue) string {
	if v == nil {
		return ""
	}
	sv, ok := v.(*types.AttributeValueMemberS)
	if !ok {
		return ""
	}
	return sv.Value
}

func sha256Hex(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}
