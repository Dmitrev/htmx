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
	_ "github.com/go-sql-driver/mysql"
)

func (t *Transaction) ToMoney() string {
    return formatMoney(t.Amount)
}

func formatMoney(money int) string {
    amount := float64(money / 100)
    formatted := strconv.FormatFloat(amount, 'f', 2, 64)

    return fmt.Sprintf("£%s", formatted)
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
    d, err := sql.Open("mysql", "user:pass@tcp(localhost:3306)/database")
    db = d
    
    check(err)

    err = db.Ping()
    check(err)

    http.HandleFunc("/", getRoot)  
    http.HandleFunc("/transactions", getTransactions)  
    http.HandleFunc("/store", postStore)  
    http.HandleFunc("/component-transactions", getComponentTransactions)  
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
    logRequest(r)
    if r.URL.Path != "/" {
	http.NotFound(w, r)
	return
    }


    tmpl := template.Must(template.ParseFiles("html/index.html", "html/partials/index.html"))

    nav := getNav(r.URL.Path)
    data := PageData{"Home", nav, nil, nil}
    err := tmpl.ExecuteTemplate(w, "index", data)
    check(err)
}

func getTransactions(w http.ResponseWriter, r *http.Request) {
    logRequest(r)
    tmpl := template.Must(template.ParseFiles("html/index.html", "html/partials/create-transaction.html"))

    nav := getNav(r.URL.Path)
    data := PageData{"Transactions", nav, nil, nil}
    err := tmpl.ExecuteTemplate(w, "index", data)
    check(err)
}

func getClicked(w http.ResponseWriter, r *http.Request) {
    logRequest(r)
    serveResponse(w, time.Now().Format(time.DateTime))
}

func getHello(w http.ResponseWriter, r *http.Request) {
    serveResponse(w, "Hello, HTTP! \n")
    logRequest(r)
}


func getNav(path string) Nav {
    return Nav {
	Items: []*NavItem {
	    {"Home", "/", path == "/"},
	    {"Transactions", "/transactions", path == "/transactions"},
	},
    }
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
	nav := getNav(r.URL.Path)
	data := PageData{"Page", nav, errors, nil}
	err := tmpl.ExecuteTemplate(w, "content", data)

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
    err = tmpl.ExecuteTemplate(w, "content", nil)
    check(err)
}

func truncate(w http.ResponseWriter, r *http.Request) {
    if r.Method != "DELETE" {
	errMethodNotAllowed(w)
	return
    }

    _, err := db.Exec("DELETE FROM transactions")
    check(err)
    w.Header().Add("HX-Trigger", "new-transactions")
    serveResponse(w, "ok")
}

func postImport(w http.ResponseWriter, r *http.Request) {
    logRequest(r)
    if r.Method != "POST" {
	errMethodNotAllowed(w)
	return
    }

    err := r.ParseMultipartForm(100 * 1024 * 1024)
    if err != nil {
	errBadRequest(w, err.Error())
	return
    }

    file, fileHeader, err := r.FormFile("file")
    if err != nil {
	errBadRequest(w, err.Error())
	return
    }

    bytes := make([]byte, fileHeader.Size)

    _, err = file.Read(bytes)
    fileContent := string(bytes[:])

    // transactions, err := ReadTransactions(fileContent)
    transactions, err := ReadFromCSV(fileContent)
    check(err)
	//
    for _, transaction := range transactions {
	// Check if exists
	stmt, err := db.Prepare(`
	    SELECT COUNT(*) FROM transactions
	    WHERE external_transaction_id = ?
	`)
	check(err)

	date := transaction.Date.Format(time.DateOnly)
	row := stmt.QueryRow(transaction.TransctionId)
	var count int
	row.Scan(&count)

	if (count > 0) {
	    fmt.Printf("%#v", count)
	    continue;
	}

	stmt, err = db.Prepare(
	    `INSERT INTO transactions (
		amount,
                date,
                description,
                payee,
                address,
                category,
                external_transaction_id
	    ) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	check(err)

	_, err = stmt.Exec(
	    transaction.Amount,
	    date,
	    transaction.Memo,
	    transaction.Payee,
	    transaction.Address,
	    transaction.Category,
	    transaction.TransctionId,
	)
	check(err)

	err = stmt.Close()
	check(err)
    }
    tmpl := template.Must(template.ParseFiles("html/partials/create-transaction.html"))

    // If has errors return form with errors
    nav := getNav(r.URL.Path)
    messages := make(map[string]string)
    messages["import"] = "Sucessfully imported";

    w.Header().Add("HX-Trigger", "new-transactions")

    data := PageData{"Page", nav, nil, messages}
    err = tmpl.ExecuteTemplate(w, "content", data)

    check(err)
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

func getComponentTransactions(w http.ResponseWriter, r *http.Request) {
    // logRequest(r)
    rows, err := db.Query(`SELECT 
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

    check(err)
    
    transactions := make([]*Transaction, 0)
    total := 0

    for rows.Next() {
	var id int
	var amount int
	var date, description string
	var payee, address, category, createdAt, updatedAt sql.NullString

	err := rows.Scan(&id, &amount, &date, &description, &payee, &address, &category, &createdAt, &updatedAt)
	if err != nil {
	    check(err)
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

    totalFormatted := formatMoney(total)

    tmpl := template.Must(template.ParseFiles("html/partials/transactions.html"))
    data := TransactionData{totalFormatted, transactions}

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
