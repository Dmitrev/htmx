package database

import (
	"database/sql"
	"fmt"
	"strconv"
)

type Transaction struct {
    Id int
    Amount int
    Date string
    Description string
    Payee string
    Address string
    Category string
    CreatedAt string
    UpdatedAt string
}

func (t *Transaction) ToMoney() string {

    amount := float64(t.Amount / 100)
    formatted := strconv.FormatFloat(amount, 'f', 2, 64)

    return fmt.Sprintf("£%s", formatted)
}

type TransactionRepo struct {
    db *sql.DB
}

func MakeTransactionRepo(db *sql.DB) TransactionRepo {
    return TransactionRepo{db: db}
}

func (t *TransactionRepo) GetAllTransactions() ([]*Transaction, error) {
    rows, err := t.db.Query(`SELECT 
	    id,
	    amount,
	    date,
	    description,
	    payee,
	    address,
	    category,
	    created_at,
	    updated_at
	FROM transactions order by date DESC`) 


    if err != nil {
	return nil, err
    }

    transactions := make([]*Transaction, 0)

    total := 0

    for rows.Next() {
	var id int
	var amount int
	var date, description string
	var payee, address, category, createdAt, updatedAt sql.NullString

	err := rows.Scan(&id, &amount, &date, &description, &payee, &address, &category, &createdAt, &updatedAt)
	if err != nil {
	    return nil, err
	}

	total += amount


	createdAtString := ""
	if createdAt.Valid {
	    createdAtString = createdAt.String
	}

	updatedAtString := ""
	if updatedAt.Valid {
	    updatedAtString = createdAt.String
	}

	payeeString := ""
	if payee.Valid {
	    payeeString = payee.String
	}

	addressString := ""
	if address.Valid {
	    addressString = address.String
	}

	categoryString := ""
	if category.Valid {
	    categoryString = category.String
	}

	t := &Transaction{
	    id, 
	    amount, 
	    date, 
	    description,
	    payeeString,
	    addressString,
	    categoryString,
	    createdAtString,
	    updatedAtString,
	}

	transactions = append(transactions, t)
    }

    return transactions, nil
}
