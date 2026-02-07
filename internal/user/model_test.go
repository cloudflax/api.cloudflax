package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_SetPassword(t *testing.T) {
	u := &User{}
	err := u.SetPassword("mypassword123")
	require.NoError(t, err)
	assert.NotEmpty(t, u.PasswordHash)
	assert.NotEqual(t, "mypassword123", u.PasswordHash)
}

func TestUser_CheckPassword(t *testing.T) {
	u := &User{}
	require.NoError(t, u.SetPassword("mypassword123"))

	assert.True(t, u.CheckPassword("mypassword123"))
	assert.False(t, u.CheckPassword("wrongpassword"))
	assert.False(t, u.CheckPassword(""))
}
