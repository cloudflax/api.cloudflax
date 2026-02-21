package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSetPassword(test *testing.T) {
	user := &User{}
	err := user.SetPassword("mypassword123")
	require.NoError(test, err)
	assert.NotEmpty(test, user.PasswordHash)
	assert.NotEqual(test, "mypassword123", user.PasswordHash)
}

func TestUserCheckPassword(test *testing.T) {
	user := &User{}
	require.NoError(test, user.SetPassword("mypassword123"))

	assert.True(test, user.CheckPassword("mypassword123"))
	assert.False(test, user.CheckPassword("wrongpassword"))
	assert.False(test, user.CheckPassword(""))
}
