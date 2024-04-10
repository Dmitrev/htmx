package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)


type QifTransaction struct {
    Date time.Time
    Amount int64
    Payee string
    Category string
    Memo string
    Address string
}

func ReadTransactions(contents string) ([]QifTransaction, error) {
    transactions := []QifTransaction{}
    
    transaction := QifTransaction{}
    for index, line := range strings.Split(contents, "\n") {

	if index == 0 {
	    continue
	}

	// Date
	if (strings.HasPrefix(line, "D")) {
	    date, err := time.Parse("02/01/2006", line[1:])
	    check(err)
	    transaction.Date = date
	    continue
	}

	// Amount
	if (strings.HasPrefix(line, "T")) {
	    amountFloat, err := strconv.ParseFloat(line[1:], 64)

	    if err != nil {
		return nil, err
	    }

	    amountFloat *= 100
	    amountInt := int64(amountFloat)

	    transaction.Amount = amountInt
	    continue
	}

	// Payee
	if (strings.HasPrefix(line, "P")) {
	    transaction.Payee = line[1:]
	    continue
	}

	// Category
	if (strings.HasPrefix(line, "L")) {
	    transaction.Category = line[1:]
	    continue
	}

	// Memo
	if (strings.HasPrefix(line, "M")) {
	    transaction.Memo = line[1:]
	    continue
	}

	// Address
	if (strings.HasPrefix(line, "A")) {
	    transaction.Address = line[1:]
	    continue
	}

	// Line ending detected
	if strings.Trim(line, " ") == "^" {
	    transactions = append(transactions, transaction)
	    transaction = QifTransaction{}
	    continue
	}

	// file end
	if line == "" {
	    break
	}
	
	return nil, fmt.Errorf("unexpected line: '%s'", line)
    }

    return transactions, nil
}
