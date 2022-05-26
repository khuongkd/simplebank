package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/khuongkd/simplebank/util"
	"github.com/stretchr/testify/require"
)

// createTestAccount create new account
func createTestAccount(t *testing.T) Account {
	arg := CreateAcountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAcount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

// deleteAccount by ID
func deleteTestAccount(t *testing.T, ID int64) {
	err := testQueries.DeleteAccount(context.Background(), ID)
	require.NoError(t, err)
	account, err := testQueries.GetAccount(context.Background(), ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, account)
}

func TestCreateAndDeleteAccount(t *testing.T) {
	account := createTestAccount(t)
	deleteTestAccount(t, account.ID)
}

func TestUpdateAccount(t *testing.T) {
	account := createTestAccount(t)
	arg := UpdateAccountParams{
		Balance: util.RandomMoney(),
		ID:      account.ID,
	}
	updatedAccount, err := testQueries.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount)
	require.Equal(t, updatedAccount.Owner, account.Owner)
	require.Equal(t, updatedAccount.Balance, arg.Balance)
	require.Equal(t, updatedAccount.Currency, account.Currency)
	require.WithinDuration(t, updatedAccount.CreatedAt, account.CreatedAt, time.Second)

	// deleteTestAccount(t, account.ID)
}

func TestGetAccount(t *testing.T) {
	account1 := createTestAccount(t)
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account2)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
	// deleteTestAccount(t, account1.ID)
}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createTestAccount(t)
	}
	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}
	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, account := range accounts {
		require.NotEmpty(t, account)
	}
}
