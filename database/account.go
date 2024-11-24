package database

import (
	"database/sql"
	"fmt"
)

type Account struct {
    Id int64 
    Name string
}

type AccountRepo struct {
    db *sql.DB
}

func MakeAccountRepo(db *sql.DB) AccountRepo {
    return AccountRepo{db: db}
}

func (r *AccountRepo) CreateAccount(name string) (*Account, error) {
    stmt, err := r.db.Prepare("INSERT INTO accounts (name) VALUES (?)")

    if err != nil {
	return nil, err
    }

    result, err := stmt.Exec(name)

    if err != nil {
	return nil, err
    }

    lastId, err := result.LastInsertId()

    if err != nil {
	return nil, err
    }

    fmt.Printf("Last inserted id: %d", lastId)

    return &Account{
	Id: lastId,
	Name: name,
    }, nil

}

func (r *AccountRepo) GetFirstAccount() Account {
    row := r.db.QueryRow("select id, name from accounts limit 1")
   
    var id int64
    var name string
    row.Scan(&id, &name)
    account := Account{Id: id, Name: name}

    return account 
}

func (r *AccountRepo) GetAllAccounts() ([]*Account, error) {
    rows, err := r.db.Query("select id, name from accounts")

    accounts := make([]*Account, 0)

    if err != nil {
	return accounts, err
    }

    for rows.Next() {
	var id int64
	var name string

	if err := rows.Scan(&id, &name); err != nil {
	    fmt.Println("Failed to scan row, skipping")
	    continue
	}
    
	accounts = append(accounts, &Account{id, name})
    }

    return accounts, nil
}
