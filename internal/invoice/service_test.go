package invoice

import (
	"testing"

	"github.com/cloudflax/api.cloudflax/internal/account"
	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupServiceTest(t *testing.T) (*Service, *account.Account) {
	t.Helper()
	require.NoError(t, database.InitForTesting())
	require.NoError(t, database.RunMigrations(&user.User{}, &account.Account{}, &account.AccountMember{}, &Invoice{}))

	acc := &account.Account{Name: "Acme", Slug: "acme"}
	require.NoError(t, database.DB.Create(acc).Error)

	repository := NewRepository(database.DB)
	return NewService(repository), acc
}

func TestService_CreateInvoice_Success(t *testing.T) {
	service, acc := setupServiceTest(t)

	inv, err := service.CreateInvoice(acc.ID, "INV-001", "USD", 9900)
	require.NoError(t, err)
	assert.NotEmpty(t, inv.ID)
	assert.Equal(t, acc.ID, inv.AccountID)
	assert.Equal(t, "INV-001", inv.Number)
	assert.Equal(t, StatusDraft, inv.Status)
	assert.Equal(t, int64(9900), inv.TotalCents)
	assert.Equal(t, "USD", inv.Currency)
}

func TestService_ListInvoice_Success(t *testing.T) {
	service, acc := setupServiceTest(t)

	_, err := service.CreateInvoice(acc.ID, "INV-001", "USD", 1000)
	require.NoError(t, err)
	_, err = service.CreateInvoice(acc.ID, "INV-002", "EUR", 2000)
	require.NoError(t, err)

	invoices, err := service.ListInvoice(acc.ID)
	require.NoError(t, err)
	assert.Len(t, invoices, 2)
}

func TestService_GetInvoice_Success(t *testing.T) {
	service, acc := setupServiceTest(t)

	created, err := service.CreateInvoice(acc.ID, "INV-001", "USD", 5000)
	require.NoError(t, err)

	found, err := service.GetInvoice(created.ID, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestService_GetInvoice_NotFound(t *testing.T) {
	service, acc := setupServiceTest(t)

	_, err := service.GetInvoice("00000000-0000-0000-0000-000000000000", acc.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestService_GetInvoice_WrongAccount(t *testing.T) {
	service, acc := setupServiceTest(t)

	otherAcc := &account.Account{Name: "Other", Slug: "other"}
	require.NoError(t, database.DB.Create(otherAcc).Error)

	created, err := service.CreateInvoice(acc.ID, "INV-001", "USD", 5000)
	require.NoError(t, err)

	_, err = service.GetInvoice(created.ID, otherAcc.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}
