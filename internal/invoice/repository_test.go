package invoice

import (
	"testing"

	"github.com/cloudflax/api.cloudflax/internal/account"
	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRepositoryTest(t *testing.T) *Repository {
	t.Helper()
	require.NoError(t, database.InitForTesting())
	require.NoError(t, database.RunMigrations(&user.User{}, &account.Account{}, &account.AccountMember{}, &Invoice{}))
	return NewRepository(database.DB)
}

func seedAccount(t *testing.T, name, slug string) *account.Account {
	t.Helper()
	acc := &account.Account{Name: name, Slug: slug}
	require.NoError(t, database.DB.Create(acc).Error)
	return acc
}

func seedInvoice(t *testing.T, repository *Repository, accountID, number string) *Invoice {
	t.Helper()
	inv := &Invoice{
		AccountID:  accountID,
		Number:     number,
		Status:     StatusDraft,
		TotalCents: 1000,
		Currency:   "USD",
	}
	require.NoError(t, repository.CreateInvoice(inv))
	return inv
}

func TestRepository_CreateInvoice_Success(t *testing.T) {
	repository := setupRepositoryTest(t)
	acc := seedAccount(t, "Acme", "acme")

	inv := &Invoice{
		AccountID:  acc.ID,
		Number:     "INV-001",
		Status:     StatusDraft,
		TotalCents: 5000,
		Currency:   "EUR",
	}
	err := repository.CreateInvoice(inv)
	require.NoError(t, err)
	assert.NotEmpty(t, inv.ID)
	assert.Equal(t, acc.ID, inv.AccountID)
}

func TestRepository_GetInvoice_Success(t *testing.T) {
	repository := setupRepositoryTest(t)
	acc := seedAccount(t, "Acme", "acme")
	created := seedInvoice(t, repository, acc.ID, "INV-001")

	found, err := repository.GetInvoice(created.ID, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, acc.ID, found.AccountID)
}

func TestRepository_GetInvoice_WrongAccount_NotFound(t *testing.T) {
	repository := setupRepositoryTest(t)
	acc1 := seedAccount(t, "Acme", "acme")
	acc2 := seedAccount(t, "Beta", "beta")
	inv := seedInvoice(t, repository, acc1.ID, "INV-001")

	_, err := repository.GetInvoice(inv.ID, acc2.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRepository_GetInvoice_NotFound(t *testing.T) {
	repository := setupRepositoryTest(t)
	acc := seedAccount(t, "Acme", "acme")

	_, err := repository.GetInvoice("00000000-0000-0000-0000-000000000000", acc.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRepository_ListInvoice_ReturnsOnlyAccountInvoices(t *testing.T) {
	repository := setupRepositoryTest(t)
	acc1 := seedAccount(t, "Acme", "acme")
	acc2 := seedAccount(t, "Beta", "beta")

	seedInvoice(t, repository, acc1.ID, "INV-001")
	seedInvoice(t, repository, acc1.ID, "INV-002")
	seedInvoice(t, repository, acc2.ID, "INV-100")

	invoices, err := repository.ListInvoice(acc1.ID)
	require.NoError(t, err)
	assert.Len(t, invoices, 2)
	for _, inv := range invoices {
		assert.Equal(t, acc1.ID, inv.AccountID)
	}
}

func TestRepository_ListInvoice_EmptyList(t *testing.T) {
	repository := setupRepositoryTest(t)
	acc := seedAccount(t, "Empty", "empty")

	invoices, err := repository.ListInvoice(acc.ID)
	require.NoError(t, err)
	assert.Empty(t, invoices)
}
