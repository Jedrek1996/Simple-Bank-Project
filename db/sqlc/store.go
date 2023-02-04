package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Composition, extend struct functionallity instead of inheritance
// By embedding queries inside store all individual queries will be avaiable to store

// Store - A batch of SQL code that can be used over and over again, dont need to keep calling everyt
type Store struct {
	*Queries
	db *sql.DB //Required to create a db transaction
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// Create a db transaction, create a new query and call the callback functions of the transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {

	//New transaction
	tx, err := store.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	q := New(tx) //New query

	err = fn(q)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction (tx) err:%v, rollback error:%v", err, rbErr)
		}
		return err
	}
	return tx.Commit() //Commit transaction``
}

// Input parameters of the transfer transaction
type TransferTxParam struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}
//Result of the transrer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTX performes a money transfer from one account to the other.
// Creates a transfer record, add account entires and update accounts balance within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParam) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {

		var err error
		//Transfer record
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})

		if err != nil {
			return err
		}
		//From account entry
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})

		if err != nil {
			return err
		}

		//To account rntry
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})

		if err != nil {
			return err
		}

		//update accounts balance

		account1, err := q.GetAccount(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.FromAccountID,
			Balance: account1.Balance - arg.Amount,
		})

		return nil
	})
	return result, err
}
