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

const (
	loginWindowSeconds    = int64(10 * 60)
	loginMaxAttempts      = int64(30)
	loginLockSeconds      = int64(30 * 60)
	loginStateTTLBuffer   = int64(24 * 60 * 60)

	refreshWindowSeconds   = int64(10 * 60)
	refreshMaxAttempts   = int64(200)
	refreshLockSeconds   = int64(15 * 60)
	refreshStateTTLBuffer = int64(24 * 60 * 60)
)

// En: ErrIPThrottleRateLimited is returned when login or refresh is throttled by IP.
// Es: ErrIPThrottleRateLimited se devuelve cuando login o refresh superan el limite por IP.
var ErrIPThrottleRateLimited = errors.New("auth ip throttle rate limited")

// En: IPThrottleGuard rate-limits auth endpoints by client IP (DynamoDB-backed).
// Es: IPThrottleGuard limita por IP endpoints de auth respaldado en DynamoDB.
type IPThrottleGuard interface {
	CheckAndConsume(ctx context.Context, ip string) error
}

// En: IPThrottleRateLimitError carries Retry-After for throttled auth-by-IP responses.
// Es: IPThrottleRateLimitError incluye Retry-After para respuestas 429 por IP.
type IPThrottleRateLimitError struct {
	RetryAfter time.Duration
}

// En: Error implements error.
func (e *IPThrottleRateLimitError) Error() string {
	return ErrIPThrottleRateLimited.Error()
}

// En: Is supports errors.Is for ErrIPThrottleRateLimited.
func (e *IPThrottleRateLimitError) Is(target error) bool {
	return target == ErrIPThrottleRateLimited
}

// En: DynamoIPThrottleGuardOptions configures a DynamoDB IP throttle guard.
// Es: DynamoIPThrottleGuardOptions configura el guard de throttle por IP en DynamoDB.
type DynamoIPThrottleGuardOptions struct {
	TableName       string
	EndpointURL     string
	Region          string
	Profile         string
	AccessKeyID     string
	SecretAccessKey string
}

type dynamoIPThrottleGuard struct {
	client    dynamoAPI
	tableName string
	now       func() time.Time
	pkPrefix  string
	maxAttempts,
	windowSeconds,
	lockSeconds,
	ttlBuffer int64
}

// En: NewDynamoLoginIPThrottleGuard creates a Dynamo-backed throttle for POST /auth/login by IP.
// Es: NewDynamoLoginIPThrottleGuard crea throttle por IP para login en DynamoDB.
func NewDynamoLoginIPThrottleGuard(ctx context.Context, opts DynamoIPThrottleGuardOptions) (IPThrottleGuard, error) {
	return newDynamoIPThrottleGuard(ctx, opts, "LOGIN", loginMaxAttempts, loginWindowSeconds, loginLockSeconds, loginStateTTLBuffer)
}

// En: NewDynamoRefreshIPThrottleGuard creates a Dynamo-backed throttle for POST /auth/refresh by IP.
// Es: NewDynamoRefreshIPThrottleGuard crea throttle por IP para refresh en DynamoDB.
func NewDynamoRefreshIPThrottleGuard(ctx context.Context, opts DynamoIPThrottleGuardOptions) (IPThrottleGuard, error) {
	return newDynamoIPThrottleGuard(ctx, opts, "REFRESH", refreshMaxAttempts, refreshWindowSeconds, refreshLockSeconds, refreshStateTTLBuffer)
}

func newDynamoIPThrottleGuard(
	ctx context.Context,
	opts DynamoIPThrottleGuardOptions,
	kind string,
	maxAttempts, windowSeconds, lockSeconds, ttlBuffer int64,
) (IPThrottleGuard, error) {
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
		return nil, fmt.Errorf("create dynamodb client for ip throttle guard: %w", err)
	}

	return &dynamoIPThrottleGuard{
		client:        client,
		tableName:     tableName,
		now:          time.Now,
		pkPrefix:     fmt.Sprintf("THROTTLE#%s#IP", kind),
		maxAttempts:  maxAttempts,
		windowSeconds: windowSeconds,
		lockSeconds:  lockSeconds,
		ttlBuffer:    ttlBuffer,
	}, nil
}

func (g *dynamoIPThrottleGuard) CheckAndConsume(ctx context.Context, ip string) error {
	normalizedIP := strings.TrimSpace(ip)
	if normalizedIP == "" {
		normalizedIP = "unknown"
	}
	ipHash := sha256Hex(normalizedIP)
	pk := fmt.Sprintf("%s#%s", g.pkPrefix, ipHash)

	const maxRetries = 4
	for range maxRetries {
		now := g.now().Unix()
		current, err := g.getState(ctx, pk)
		if err != nil {
			return fmt.Errorf("load ip throttle state: %w", err)
		}

		next, limitErr := evaluateIPWindowState(
			current, now, g.maxAttempts, g.windowSeconds, g.lockSeconds, g.ttlBuffer,
		)
		if next == nil && limitErr != nil {
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
			return fmt.Errorf("store ip throttle state: %w", err)
		}
		if limitErr != nil {
			return limitErr
		}
		return nil
	}

	return fmt.Errorf("store ip throttle state: too much contention")
}

func evaluateIPWindowState(
	current *resendGuardState,
	now int64,
	maxAttempts, windowSeconds, lockSeconds, ttlBuffer int64,
) (*resendGuardState, *IPThrottleRateLimitError) {
	if current != nil && current.Exists && current.LockUntil > now {
		return nil, &IPThrottleRateLimitError{
			RetryAfter: time.Duration(current.LockUntil-now) * time.Second,
		}
	}

	if current != nil && current.Exists && current.LockUntil != 0 && current.LockUntil <= now {
		// Lock expired: new window, first attempt in that window.
		next := buildIPThrottleTransition(current, now, 1, now, 0, windowSeconds, ttlBuffer)
		return next, nil
	}

	if current == nil || !current.Exists {
		next := buildIPThrottleTransition(&resendGuardState{}, now, 1, now, 0, windowSeconds, ttlBuffer)
		return next, nil
	}

	if now >= current.WindowStart+windowSeconds {
		next := buildIPThrottleTransition(current, now, 1, now, 0, windowSeconds, ttlBuffer)
		return next, nil
	}

	newCount := current.Count + 1
	if newCount > maxAttempts {
		lockUntil := now + lockSeconds
		next := buildIPThrottleTransition(current, now, newCount, current.WindowStart, lockUntil, windowSeconds, ttlBuffer)
		return next, &IPThrottleRateLimitError{
			RetryAfter: time.Duration(lockSeconds) * time.Second,
		}
	}

	next := buildIPThrottleTransition(current, now, newCount, current.WindowStart, 0, windowSeconds, ttlBuffer)
	return next, nil
}

// buildIPThrottleTransition increments version and updates window metrics; windowStart is the start of the active window.
func buildIPThrottleTransition(
	previous *resendGuardState,
	now, count, windowStart, lockUntil, windowSeconds, ttlBuffer int64,
) *resendGuardState {
	version := int64(1)
	createdAt := now
	pk := ""
	sk := ""
	if previous != nil && previous.Exists {
		version = previous.Version + 1
		if previous.CreatedAt != 0 {
			createdAt = previous.CreatedAt
		}
		pk = previous.OriginalPK
		sk = previous.OriginalSK
	}

	expiresAt := windowStart + windowSeconds + ttlBuffer
	if lockUntil > 0 {
		expiresAt = lockUntil + ttlBuffer
	}

	return &resendGuardState{
		Exists:        true,
		OriginalPK:    pk,
		OriginalSK:    sk,
		Version:       version,
		Count:         count,
		WindowStart:   windowStart,
		NextAllowedAt: 0,
		LockUntil:     lockUntil,
		CreatedAt:     createdAt,
		UpdatedAt:     now,
		ExpiresAt:     expiresAt,
	}
}

func (g *dynamoIPThrottleGuard) getState(ctx context.Context, pk string) (*resendGuardState, error) {
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

func (g *dynamoIPThrottleGuard) putState(ctx context.Context, current, next *resendGuardState) error {
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
