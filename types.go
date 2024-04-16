package main

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

type TransactionData struct {
    Total string
    Transactions []*Transaction
}

type PageData struct {
    Title string
    Nav Nav
    Errors map[string]string
    Messages map[string]string
}

type Nav struct {
    Items []*NavItem
}

type NavItem struct {
    Label string
    Url string
    Active bool
}
