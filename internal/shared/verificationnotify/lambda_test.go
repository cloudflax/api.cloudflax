package verificationnotify

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerificationPayloadMarshal(t *testing.T) {
	t.Parallel()
	p := verificationPayload{Email: "u@x.com", Name: "U", Link: "https://app/x?token=1"}
	b, err := json.Marshal(p)
	require.NoError(t, err)
	assert.JSONEq(t, `{"email":"u@x.com","name":"U","link":"https://app/x?token=1"}`, string(b))
}

func TestNewLambdaNotifierRequiresFunctionName(t *testing.T) {
	t.Parallel()
	_, err := NewLambdaNotifier(context.Background(), LambdaNotifierOptions{Region: "us-east-1", FunctionName: ""})
	require.Error(t, err)
}

type stubLambdaClient struct {
	lastInput *lambda.InvokeInput
}

func (s *stubLambdaClient) Invoke(ctx context.Context, params *lambda.InvokeInput, _ ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	s.lastInput = params
	return &lambda.InvokeOutput{StatusCode: 202}, nil
}

func TestLambdaNotifierNotifyVerificationEmail(t *testing.T) {
	t.Parallel()
	stub := &stubLambdaClient{}
	n := &LambdaNotifier{client: stub, functionName: "my-fn"}

	err := n.NotifyVerificationEmail(context.Background(), "a@b.com", "Alice", "https://front/auth/verify-email?token=t")
	require.NoError(t, err)

	require.NotNil(t, stub.lastInput)
	assert.Equal(t, types.InvocationTypeEvent, stub.lastInput.InvocationType)
	require.NotNil(t, stub.lastInput.FunctionName)
	assert.Equal(t, "my-fn", *stub.lastInput.FunctionName)

	var got verificationPayload
	require.NoError(t, json.Unmarshal(stub.lastInput.Payload, &got))
	assert.Equal(t, "a@b.com", got.Email)
	assert.Equal(t, "Alice", got.Name)
	assert.Equal(t, "https://front/auth/verify-email?token=t", got.Link)
}

func TestLambdaNotifierNotifyVerificationEmailEmptyRecipient(t *testing.T) {
	t.Parallel()
	n := &LambdaNotifier{client: &stubLambdaClient{}, functionName: "fn"}
	err := n.NotifyVerificationEmail(context.Background(), "  ", "N", "https://x")
	assert.Error(t, err)
}

func TestPasswordResetPayloadMarshal(t *testing.T) {
	t.Parallel()
	p := passwordResetPayload{Email: "u@x.com", Name: "U", Link: "https://app/r?t=1", ExpiresIn: "60 minutes"}
	b, err := json.Marshal(p)
	require.NoError(t, err)
	assert.JSONEq(t, `{"email":"u@x.com","name":"U","link":"https://app/r?t=1","expiresIn":"60 minutes"}`, string(b))
}

func TestLambdaNotifierNotifyPasswordResetEmail(t *testing.T) {
	t.Parallel()
	stub := &stubLambdaClient{}
	n := &LambdaNotifier{client: stub, functionName: "forgot-fn"}

	err := n.NotifyPasswordResetEmail(context.Background(), "a@b.com", "Bob", "https://front/auth/reset-password?token=t", "60 minutes")
	require.NoError(t, err)

	var got passwordResetPayload
	require.NoError(t, json.Unmarshal(stub.lastInput.Payload, &got))
	assert.Equal(t, "a@b.com", got.Email)
	assert.Equal(t, "Bob", got.Name)
	assert.Equal(t, "https://front/auth/reset-password?token=t", got.Link)
	assert.Equal(t, "60 minutes", got.ExpiresIn)
}
