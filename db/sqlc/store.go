package db

import (
	"context"
	"database/sql"
	"fmt"
)

/*

BEGIN;

INSERT INTO transfers (from_account_id, to_account_id, amount) VALUES (1, 2, 5) RETURNING *;

INSERT INTO entries (account_id, amount) VALUES (1, -5) RETURNING *;
INSERT INTO entries (account_id, amount) VALUES (2, 5) RETURNING *;

SELECT * from accounts WHERE id=1 FOR NO KEY UPDATE;
UPDATE accounts SET balance = 95 where id = 1 RETURNING *;

SELECT * from accounts WHERE id=2 FOR NO KEY UPDATE;
UPDATE accounts SET balance = 105 where id = 2 RETURNING *;

ROLLBACK;

How to avoid deadlock during DB transaction
1. Handle the query within the transaction in a such way that it minimize the chance of lock of record.
2. Maintain the consistency across the same type of transaction

*/

// Store defines all functions to execute db queries and transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg CreateTransferParams) (TransferTxResult, error)
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	db *sql.DB
	*Queries
}

func NewStore(db *sql.DB) *SQLStore {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// this method would execute a function within database transaction
// ExecTx executes a function within a database transaction
// Higher Order Function or Composible Function to Transaction behaviour (Begin => Commit)
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Get the new transaction queries
	q := New(tx)
	err = fn(q)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// TransferTx performs a money transfer from one account to the other.
/*
 1. It creates the transfer
 2. add account entries
 3. update accounts balance within a database transaction
*/
// TransferTxParams contains the input parameters of the transfer transaction
// type TransferTxParams struct {
// 	FromAccountID int64 `json:"from_account_id"`
// 	ToAccountID   int64 `json:"to_account_id"`
// 	Amount        int64 `json:"amount"`
// }

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (store *SQLStore) TransferTx(ctx context.Context, arg CreateTransferParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		// Step 1: Create Transfer record
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// Step 2: Create Record Entry in the FromAccountID with negative amount (Since money is getting debited)
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// Step 3: Create Record Entry for ToAccountID with amount (Since money is getting credited)
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		return err
	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) { // Variable of the account 1 & account2 is decided inline
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	// Example of the naked return it will return that is inline to the function signature
	return
}
