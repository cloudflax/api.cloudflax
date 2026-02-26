package account

import (
	"testing"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupServiceTest(t *testing.T) *Service {
	t.Helper()
	require.NoError(t, database.InitForTesting())
	require.NoError(t, database.RunMigrations(&user.User{}, &Account{}, &AccountMember{}))

	userRepository := user.NewRepository(database.DB)
	accountRepository := NewRepository(database.DB)
	return NewService(accountRepository, userRepository)
}

func seedVerifiedUser(t *testing.T, name, email string) *user.User {
	t.Helper()
	now := time.Now()
	u := &user.User{Name: name, Email: email, EmailVerifiedAt: &now}
	require.NoError(t, u.SetPassword("password123"))
	require.NoError(t, database.DB.Create(u).Error)
	return u
}

func seedUnverifiedUser(t *testing.T, name, email string) *user.User {
	t.Helper()
	u := &user.User{Name: name, Email: email}
	require.NoError(t, u.SetPassword("password123"))
	require.NoError(t, database.DB.Create(u).Error)
	return u
}

func TestService_ListAccountsForUser_Empty(t *testing.T) {
	service := setupServiceTest(t)
	owner := seedVerifiedUser(t, "Alice", "alice.list.empty@example.com")

	accounts, err := service.ListAccountsForUser(owner.ID)
	require.NoError(t, err)
	assert.Len(t, accounts, 0)
}

func TestService_ListAccountsForUser_WithAccounts(t *testing.T) {
	service := setupServiceTest(t)
	owner := seedVerifiedUser(t, "Bob", "bob.list@example.com")

	// create two accounts for the same owner
	first, _, err := service.CreateAccount("First Org", "", owner.ID)
	require.NoError(t, err)
	second, _, err := service.CreateAccount("Second Org", "", owner.ID)
	require.NoError(t, err)

	accounts, err := service.ListAccountsForUser(owner.ID)
	require.NoError(t, err)
	assert.Len(t, accounts, 2)

	ids := []string{accounts[0].ID, accounts[1].ID}
	assert.Contains(t, ids, first.ID)
	assert.Contains(t, ids, second.ID)
}

func TestService_ListAccountsForUser_InvalidUserID(t *testing.T) {
	service := setupServiceTest(t)

	_, err := service.ListAccountsForUser("not-a-uuid")
	assert.Error(t, err)
}

func TestService_CreateAccount_Success(t *testing.T) {
	service := setupServiceTest(t)
	owner := seedVerifiedUser(t, "Alice", "alice@example.com")

	account, member, err := service.CreateAccount("Acme Corp", "", owner.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, account.ID)
	assert.Equal(t, "Acme Corp", account.Name)
	assert.Equal(t, "acme-corp", account.Slug)
	assert.Equal(t, RoleOwner, member.Role)
	assert.Equal(t, owner.ID, member.UserID)
	assert.Equal(t, account.ID, member.AccountID)

	// owner should now have this account as active
	updatedOwner, err := service.userRepository.GetUser(owner.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedOwner.ActiveAccountID)
	assert.Equal(t, account.ID, *updatedOwner.ActiveAccountID)
}

func TestService_CreateAccount_CustomSlug(t *testing.T) {
	service := setupServiceTest(t)
	owner := seedVerifiedUser(t, "Bob", "bob@example.com")

	account, _, err := service.CreateAccount("My Workspace", "my-ws", owner.ID)
	require.NoError(t, err)
	assert.Equal(t, "my-ws", account.Slug)
}

func TestService_SetActiveAccountForUser_Success(t *testing.T) {
	service := setupServiceTest(t)
	owner := seedVerifiedUser(t, "Henry", "henry@example.com")

	account, _, err := service.CreateAccount("Henry Org", "", owner.ID)
	require.NoError(t, err)

	// Clear active account so we can explicitly set it again.
	owner.ActiveAccountID = nil
	require.NoError(t, database.DB.Save(owner).Error)

	updatedUser, err := service.SetActiveAccountForUser(owner.ID, account.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedUser.ActiveAccountID)
	assert.Equal(t, account.ID, *updatedUser.ActiveAccountID)
}

func TestService_SetActiveAccountForUser_NotMember(t *testing.T) {
	service := setupServiceTest(t)
	owner := seedVerifiedUser(t, "Isaac", "isaac-owner@example.com")
	otherUser := seedVerifiedUser(t, "Jess", "jess@example.com")

	account, _, err := service.CreateAccount("Isaac Org", "", owner.ID)
	require.NoError(t, err)

	_, err = service.SetActiveAccountForUser(otherUser.ID, account.ID)
	assert.ErrorIs(t, err, ErrMemberNotFound)
}

func TestService_CreateAccount_UnverifiedEmail(t *testing.T) {
	service := setupServiceTest(t)
	owner := seedUnverifiedUser(t, "Carol", "carol@example.com")

	_, _, err := service.CreateAccount("Carol Corp", "", owner.ID)
	assert.ErrorIs(t, err, ErrUserEmailNotVerified)
}

func TestService_CreateAccount_DuplicateSlug(t *testing.T) {
	service := setupServiceTest(t)
	owner := seedVerifiedUser(t, "Dave", "dave@example.com")

	_, _, err := service.CreateAccount("First", "shared-slug", owner.ID)
	require.NoError(t, err)

	_, _, err = service.CreateAccount("Second", "shared-slug", owner.ID)
	assert.ErrorIs(t, err, ErrSlugTaken)
}

func TestService_CreateAccount_AutoSlugFromName(t *testing.T) {
	service := setupServiceTest(t)
	owner := seedVerifiedUser(t, "Eve", "eve@example.com")

	account, _, err := service.CreateAccount("Hello World!", "", owner.ID)
	require.NoError(t, err)
	assert.Equal(t, "hello-world", account.Slug)
}

func TestService_CreateAccount_InvalidUserID(t *testing.T) {
	service := setupServiceTest(t)

	_, _, err := service.CreateAccount("Bad", "", "not-a-uuid")
	assert.Error(t, err)
}
