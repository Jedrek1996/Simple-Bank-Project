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

// Result of the transrer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// {} empty struct {} empty object of that type
// var txKey = struct{}{}

// TransferTX performes a money transfer from one account to the other.
// Creates a transfer record, add account entires and update accounts balance within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParam) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {

		var err error

		//Transfer record
		// txName := ctx.Value(txKey)
		// fmt.Println(txName, "Create transfer")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})

		if err != nil {
			return err
		}
		//From account entry
		// fmt.Println(txName, "Create entry 1 - Deducting FromAccount")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})

		if err != nil {
			return err
		}

		//To account entry
		// fmt.Println(txName, "Create entry 2 - Adding into ToAccount")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})

		if err != nil {
			return err
		}

		//update accounts balance
		//Move money out of from account
		// fmt.Println(txName, "Update account 1")

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, _ = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, _ = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}
		return nil
	})
	return result, err
}

func addMoney(ctx context.Context, q *Queries, accountID1 int64, amount1 int64, accountID2 int64, amount2 int64) (account1 Account, account2 Account, err error) {

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
	return
}
