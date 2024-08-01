package database

import (
	"database/sql"
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

func (r *AccountRepo) GetFirstAccount() Account {
    row := r.db.QueryRow("select id, name from accounts limit 1")
   
    var id int64
    var name string
    row.Scan(&id, &name)
    account := Account{Id: id, Name: name}

    return account 
     
}
