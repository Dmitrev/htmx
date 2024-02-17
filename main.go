package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Transaction struct {
    Id int
    Amount int
    Date string
    Description string
    Payee string
    Address string
    Category string
}

func (t *Transaction) ToMoney() string {
    return formatMoney(t.Amount)
}

func formatMoney(money int) string {
    amount := float64(money / 100)
    formatted := strconv.FormatFloat(amount, 'f', 2, 64)

    return fmt.Sprintf("£%s", formatted)
}

type TransactionData struct {
    Title string
    Total string
    Transactions []*Transaction
}

type PageData struct {
    Errors map[string]string
}

func check(err error) {
    if err != nil {
	panic(err)
    }
}

var db *sql.DB;

func main() {

    args := os.Args[1:]

    if (len(args) > 0) {

    } else {
	startWebServer()
    }
}

func startWebServer() {
    d, err := sql.Open("sqlite3", "htmx.db")
    db = d
    
    check(err)

    err = db.Ping()
    check(err)

    http.HandleFunc("/", getRoot)  
    http.HandleFunc("/store", postStore)  
    http.HandleFunc("/transactions", getTransactions)  
    http.HandleFunc("/clicked", getClicked)
    http.HandleFunc("/hello", getHello)  
    http.HandleFunc("/delete/", deleteTransaction)  
    http.HandleFunc("/truncate", truncate)
    http.HandleFunc("/import", postImport)

    fmt.Println("Starting server on http://localhost:3333")
    http.ListenAndServe(":3333", nil)

    fmt.Println("calling db.Close()")
    defer db.Close()
}

func getRoot(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Caught in ROOT")
    logRequest(r)
    if r.URL.Path != "/" {
	http.NotFound(w, r)
	return
    }

    tmpl := template.Must(template.ParseFiles("html/index.html", "html/partials/create-transaction.html"))
    tmpl.ExecuteTemplate(w, "index", nil)
    // serveFile(w, r, "html/index.html")
}

func getClicked(w http.ResponseWriter, r *http.Request) {
    logRequest(r)
    serveResponse(w, time.Now().Format(time.DateTime))
}

func getHello(w http.ResponseWriter, r *http.Request) {
    serveResponse(w, "Hello, HTTP! \n")
    logRequest(r)
}

func postStore(w http.ResponseWriter, r *http.Request) {
    logRequest(r)

    if r.Method != "POST" {
	errMethodNotAllowed(w)
	return
    }

    r.ParseForm()
    
    tmpl := template.Must(template.ParseFiles("html/partials/create-transaction.html"))
    
    required := []string{"amount", "date", "description"}
    errors := make(map[string]string)
    for _, key := range required {
	if !r.Form.Has(key) || r.Form.Get(key) == "" {
	    errors[key] = fmt.Sprintf("The %s field is missing", key) 
	}
    }

    if len(errors) > 0 {
	fmt.Println("validation errors")
	// If has errors return form with errors
	data := PageData{errors}
	err := tmpl.ExecuteTemplate(w, "create-transaction", data)

	check(err)
	return
    }

    // otherwise continue inserting
    amount := r.Form.Get("amount")
    date := r.Form.Get("date")
    description := r.Form.Get("description")

    stmt, err := db.Prepare("insert into transactions (amount, date, description) VALUES (?, ?, ?)")

    check(err)

    _, err = stmt.Exec(amount, date, description)

    check(err)

    w.Header().Add("HX-Trigger", "new-transactions")
    err = tmpl.ExecuteTemplate(w, "create-transaction", nil)
    check(err)
}

func truncate(w http.ResponseWriter, r *http.Request) {
    if r.Method != "DELETE" {
	errMethodNotAllowed(w)
	return
    }

    _, err := db.Exec("DELETE FROM transactions")
    check(err)
    serveResponse(w, "ok")
}

func postImport(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
	errMethodNotAllowed(w)
	return
    }

    transactions, err := OpenFile("monzo.qif")
    check(err)

    for _, transaction := range transactions {
	// fmt.Printf("%#v\n", transaction)
	stmt, err := db.Prepare("INSERT INTO transactions (amount, date, description, payee, address, category) VALUES (?, ?, ?, ?, ?, ?)")
	check(err)
	
	date := transaction.Date.Format(time.DateOnly)

	_, err = stmt.Exec(transaction.Amount, date, transaction.Memo, transaction.Payee, transaction.Address, transaction.Category)
	check(err)
    }

    serveResponse(w, "ok")
}

func logRequest(r *http.Request) {
    _, err := fmt.Printf("[%s] %s %s\n", time.Now().Format(time.DateTime), r.Method, r.URL.Path)

    check(err)

    fmt.Println("--headers---")
    for key, values := range r.Header {
	fmt.Printf("%s: %v\n", key, values)		
    }

    if r.Method == "POST" {
	err := r.ParseForm(); 
	check(err)
    } 

    fmt.Println("--body---")
    for key, value := range r.PostForm {
	fmt.Printf("%s: %v\n", key, value)		
    }

    fmt.Println("-----")
}

func deleteTransaction(w http.ResponseWriter, r *http.Request) {
    logRequest(r)

    if r.Method != "DELETE" {
	errMethodNotAllowed(w)
	return
    }

    id := strings.Split(r.URL.Path, "/")[2]

    stmt, err := db.Prepare("DELETE FROM transactions WHERE id = ?")

    check(err)

    _, err = stmt.Exec(id)

    check(err)
}

func getTransactions(w http.ResponseWriter, r *http.Request) {
    // logRequest(r)
    rows, err := db.Query("SELECT * FROM transactions") 

    check(err)
    
    transactions := make([]*Transaction, 0)
    total := 0

    for rows.Next() {
	var id int
	var amount int
	var date, description, payee, address, category string

	rows.Scan(&id, &amount, &date, &description, &payee, &address, &category)

	total += amount

	t := &Transaction{id, amount, date, description, payee, address, category}
	transactions = append(transactions, t)
    }

    totalFormatted := formatMoney(total)

    tmpl := template.Must(template.ParseFiles("html/partials/transactions.html"))
    data := TransactionData{"title testing", totalFormatted, transactions}

    err = tmpl.Execute(w, data)
    check(err)

    // serveFile(w, r, "html/partials/transactions.html")
}

func serveFile(
    w http.ResponseWriter,
    r *http.Request,
    filepath string,
) {
    // w.Header().Add("Content-type", "text/html")
    http.ServeFile(w, r, filepath) 
}

func serveResponse(
    w http.ResponseWriter,
    body string,
) {
    w.Header().Add("Content-type", "text/html")
    io.WriteString(w, body)
}

func errMethodNotAllowed(w http.ResponseWriter) {
    w.WriteHeader(405)
}

func errBadRequest(w http.ResponseWriter, error string) {
    w.WriteHeader(400)
    serveResponse(w, error)
}
