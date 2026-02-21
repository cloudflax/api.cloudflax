package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserSetPassword tests the SetPassword method of the User model.
// En: Verifies that SetPassword correctly hashes the password and does not store it in plain text.
// Es: Verifica que SetPassword hashee correctamente la contraseña y no la almacene en texto plano.
func TestUserSetPassword(test *testing.T) {
	user := &User{}
	err := user.SetPassword("mypassword123")
	require.NoError(test, err)
	assert.NotEmpty(test, user.PasswordHash)
	assert.NotEqual(test, "mypassword123", user.PasswordHash)
}

// TestUserCheckPassword tests the CheckPassword method of the User model.
// En: Verifies that CheckPassword correctly validates the password against the stored hash.
// Es: Verifica que CheckPassword valide correctamente la contraseña contra el hash almacenado.
func TestUserCheckPassword(test *testing.T) {
	user := &User{}
	require.NoError(test, user.SetPassword("mypassword123"))

	assert.True(test, user.CheckPassword("mypassword123"))
	assert.False(test, user.CheckPassword("wrongpassword"))
	assert.False(test, user.CheckPassword(""))
}
