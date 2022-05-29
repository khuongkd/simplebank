package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, params CreateTransferParams) (TransferTxResult, error)
}

// Store provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	db := New(tx)
	err = fn(db)
	if err != nil {
		if errRb := tx.Rollback(); errRb != nil {
			return fmt.Errorf("txErr: %v, rbErr: %v", err, errRb)
		}

		return err
	}

	return tx.Commit()
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to another account
// It create a transfer record, add account entries, and update account's balance within a single database transaction
func (store *SQLStore) TransferTx(ctx context.Context, params CreateTransferParams) (TransferTxResult, error) {
	var result TransferTxResult
	err := store.execTx(ctx, func(*Queries) error {
		// create transfer
		transfer, err := store.CreateTransfer(ctx, params)
		if err != nil {
			return err
		}
		result.Transfer = transfer

		// create fromEntry
		fromEntry, err := store.CreateEntry(ctx, CreateEntryParams{
			AccountID: transfer.FromAccountID,
			Amount:    -transfer.Amount,
		})
		if err != nil {
			return err
		}
		result.FromEntry = fromEntry

		// create toEntry
		toEntry, err := store.CreateEntry(ctx, CreateEntryParams{
			AccountID: transfer.ToAccountID,
			Amount:    transfer.Amount,
		})
		if err != nil {
			return err
		}
		result.ToEntry = toEntry

		result.FromAccount, result.ToAccount, err = store.AddAccountBalanceOrder(ctx, params.FromAccountID, params.ToAccountID, params.Amount)

		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func (store *SQLStore) AddAccountBalanceOrder(ctx context.Context, account1ID, account2ID, amount int64) (fromAccount Account, toAccount Account, err error) {

	if account1ID < account2ID {
		// Update from Accounts' balance
		fromAccount, err = store.AddAccountBalance(ctx, AddAccountBalanceParams{
			Amount: -amount,
			ID:     account1ID,
		})
		if err != nil {
			return
		}

		// Update to Accounts' balance
		toAccount, err = store.AddAccountBalance(ctx, AddAccountBalanceParams{
			Amount: amount,
			ID:     account2ID,
		})
		if err != nil {
			return
		}

		return
	}

	// Update to Accounts' balance
	toAccount, err = store.AddAccountBalance(ctx, AddAccountBalanceParams{
		Amount: amount,
		ID:     account2ID,
	})
	if err != nil {
		return
	}

	// Update from Accounts' balance
	fromAccount, err = store.AddAccountBalance(ctx, AddAccountBalanceParams{
		Amount: -amount,
		ID:     account1ID,
	})
	if err != nil {
		return
	}

	return
}
