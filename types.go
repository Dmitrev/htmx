package main

import "htmx/database"

type TransactionData struct {
    Total string
    Transactions []*database.Transaction
}

type AccountData struct {
    Accounts []*database.Account
}

type PageData struct {
    Title string
    Nav Nav
    Errors map[string]string
    Values map[string]string
    Messages map[string]string
    Content any
}

type Nav struct {
    Items []*NavItem
}

type NavItem struct {
    Label string
    Url string
    Active bool
}
