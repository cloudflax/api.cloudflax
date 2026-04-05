package auth

import (
	"context"
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

// En: ErrForgotPasswordRateLimited indicates forgot-password email throttle limits were reached.
// Es: ErrForgotPasswordRateLimited indica que se alcanzaron los limites de throttle de correo de recuperacion.
var ErrForgotPasswordRateLimited = errors.New("forgot password email rate limited")

// En: ForgotPasswordGuard enforces forgot-password email throttling (same policy as resend verification).
// Es: ForgotPasswordGuard aplica throttling al correo de recuperacion de contraseña (misma politica que reenvio de verificacion).
type ForgotPasswordGuard interface {
	CheckAndConsume(ctx context.Context, email, ip string) error
}

// En: DynamoForgotPasswordGuardOptions configures the DynamoDB forgot-password guard.
// Es: DynamoForgotPasswordGuardOptions configura el guard de forgot-password en DynamoDB.
type DynamoForgotPasswordGuardOptions struct {
	TableName       string
	EndpointURL     string
	Region          string
	Profile         string
	AccessKeyID     string
	SecretAccessKey string
}

// En: ForgotPasswordRateLimitError includes retry delay for throttled forgot-password requests.
// Es: ForgotPasswordRateLimitError incluye demora de reintento cuando forgot-password esta bloqueado.
type ForgotPasswordRateLimitError struct {
	RetryAfter time.Duration
}

// En: Error returns the throttle reason message.
// Es: Error devuelve el mensaje del motivo del throttle.
func (e *ForgotPasswordRateLimitError) Error() string {
	return ErrForgotPasswordRateLimited.Error()
}

// En: Is reports whether err is a forgot-password rate-limit error.
// Es: Is indica si err es un error de limite de forgot-password.
func (e *ForgotPasswordRateLimitError) Is(target error) bool {
	return target == ErrForgotPasswordRateLimited
}

// En: dynamoForgotPasswordGuard stores throttle state in DynamoDB under a distinct pk prefix from resend verification.
// Es: dynamoForgotPasswordGuard guarda estado de throttle en DynamoDB con prefijo pk distinto al reenvio de verificacion.
type dynamoForgotPasswordGuard struct {
	client    dynamoAPI
	tableName string
	now       func() time.Time
}

// En: NewDynamoForgotPasswordGuard builds a DynamoDB-backed forgot-password guard.
// Es: NewDynamoForgotPasswordGuard crea un guard de forgot-password con DynamoDB.
func NewDynamoForgotPasswordGuard(ctx context.Context, opts DynamoForgotPasswordGuardOptions) (ForgotPasswordGuard, error) {
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
		return nil, fmt.Errorf("create dynamodb client for forgot-password guard: %w", err)
	}

	return &dynamoForgotPasswordGuard{
		client:    client,
		tableName: tableName,
		now:       time.Now,
	}, nil
}

// En: CheckAndConsume validates limits and consumes one forgot-password email quota atomically.
// Es: CheckAndConsume valida limites y consume una cuota de correo de recuperacion de forma atomica.
func (g *dynamoForgotPasswordGuard) CheckAndConsume(ctx context.Context, email, _ string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" {
		return nil
	}

	emailHash := sha256Hex(normalizedEmail)
	pk := fmt.Sprintf("THROTTLE#FORGOT_PASSWORD#EMAIL#%s", emailHash)

	const maxRetries = 4
	for range maxRetries {
		now := g.now().Unix()
		current, err := g.getState(ctx, pk)
		if err != nil {
			return fmt.Errorf("load forgot-password throttle state: %w", err)
		}

		next, limitErr := evaluateForgotPasswordState(current, now)
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
			return fmt.Errorf("store forgot-password throttle state: %w", err)
		}
		return nil
	}

	return fmt.Errorf("store forgot-password throttle state: too much contention")
}

func (g *dynamoForgotPasswordGuard) getState(ctx context.Context, pk string) (*resendGuardState, error) {
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

func (g *dynamoForgotPasswordGuard) putState(ctx context.Context, current, next *resendGuardState) error {
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

func evaluateForgotPasswordState(current *resendGuardState, now int64) (*resendGuardState, error) {
	if current == nil || !current.Exists {
		return buildFreshState(now), nil
	}

	if current.LockUntil > now {
		return nil, &ForgotPasswordRateLimitError{
			RetryAfter: time.Duration(current.LockUntil-now) * time.Second,
		}
	}

	if current.Count >= resendMaxSends {
		return buildFreshStateFrom(current, now), nil
	}

	if current.NextAllowedAt > now {
		return nil, &ForgotPasswordRateLimitError{
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
