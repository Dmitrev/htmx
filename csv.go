package main

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"time"
)

type CsvTransaction struct {
    TransctionId string
    Date time.Time
    Amount int64
    Payee string
    Category string
    Memo string
    Address string
}

func ReadFromCSV(contents string) ([]CsvTransaction, error) {
    transactions := []CsvTransaction{}

    r := csv.NewReader(strings.NewReader(contents))
    r.Read() //skip first line
    for {
	transaction := CsvTransaction{}
	
	line, err := r.Read()

	if err != nil {
	    if err == io.EOF {
		break
	    } else {
		panic(err) 
	    }
	}

	transaction.TransctionId = line[0]
	date, err := time.Parse("02/01/2006", line[1])
	panicOnErr(err)
	transaction.Date = date
	transaction.Memo = line[11]
	transaction.Category = line[6]
	transaction.Payee = line[4]
	transaction.Address = line[12]
	amount, err := strconv.ParseFloat(line[7], 64)
	panicOnErr(err)
	amount *= 100
	transaction.Amount = int64(amount)

	transactions = append(transactions, transaction)
    }
    
    return transactions, nil
}
