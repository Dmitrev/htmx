package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"htmx/components"
	"htmx/database"
	"htmx/templates"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const maxMemoryFormInBytes = 100 * 1024 * 1024


var renderer *templates.Renderer

func panicOnErr(err error) {
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


    panicOnErr(err)

    err = db.Ping()
    panicOnErr(err)
    router := CreateRouter()
    router.Get("/", getRoot)
    router.Get("/transactions", getTransactions)
    router.Get("/accounts", getAccounts)
    router.Get("/accounts/:id", showAccount)
    router.Delete("/accounts/:id", deleteAccount)
    router.Get("/component/accounts", getAccountsComponent)
    router.Post("/accounts", createAccount)
    // router.Get("/component-transactions", getComponentTransactions)
    router.Post("/store", postStore)
    router.Delete("/delete/:id", deleteTransaction)
    router.Post("/truncate", truncate)
    router.Post("/import", postImport)

    renderer = templates.MakeRenderer(true)
    startServer("localhost", 3333, router)

    //
    // fmt.Println("Starting server on http://localhost:3333")
    // http.ListenAndServe(":3333", nil)
    //
    // fmt.Println("calling db.Close()")
    // defer db.Close()



    // new router setup
    // registerRoutes("localhost", 3333)
}

func getRoot(w http.ResponseWriter, r RequestContext) {

    nav := getNav(r.Request.URL.Path)
    data := PageData{"Home", nav, nil, nil, nil}
    renderer.Render(w, "index.gohtml", data)
}

func getTransactions(w http.ResponseWriter, r RequestContext) {
    repo := database.MakeTransactionRepo(db)
    accountRepo := database.MakeAccountRepo(db)
    transactions, err := repo.GetAllTransactions()

    panicOnErr(err)
    accounts, err := accountRepo.GetAllAccounts()

    nav := getNav(r.Request.URL.Path)
    data := PageData{
	"Transactions",
	nav,
	nil,
	nil,
	struct {
	    Transactions []*database.Transaction
	    Accounts []*database.Account
	} {
	    Transactions: transactions,
	    Accounts: accounts,
	},
    }

    renderer.Render(w, "transactions.gohtml", data)
}

func getAccounts(w http.ResponseWriter, r RequestContext) {

    nav := getNav(r.Request.URL.Path)
    data := PageData{"Accounts", nav, nil, nil, nil}

    renderer.Render(w, "accounts.gohtml", data);
}

func getAccountsComponent(w http.ResponseWriter, r RequestContext) {
    type templateData struct {
	Accounts []*database.Account
	CreateButton components.Button
    }
    repo := database.MakeAccountRepo(db)
    accounts, err := repo.GetAllAccounts()

    if err != nil {
	fmt.Printf("Failed to fetch accounts, error: %s\n", err)
    }

    tmpl := template.Must(template.ParseFiles("html/partials/accounts-list.gohtml"))
    data := templateData {
	Accounts: accounts,
	CreateButton: components.Button {
	    Value: "Create Account",
	},
    }

    err = tmpl.Execute(w, data)
    panicOnErr(err)
}

func createAccount(w http.ResponseWriter, r RequestContext) {
    r.Request.ParseForm()

    errors := make(map[string]string)

    if (!r.Request.Form.Has("name") || r.Request.Form.Get("name") == "") {
	errors["name"] = "Missing name"
    }

    if len(errors) == 0 {
	repo := database.MakeAccountRepo(db)
	account, err := repo.CreateAccount(r.Request.Form.Get("name"))

	panicOnErr(err)

	fmt.Printf("%#v\n", account)
    }

    if len(errors) == 0 {
	w.Header().Add("HX-Trigger", "new-accounts")
    }

    nav := getNav(r.Request.URL.Path)
    data := PageData{"Page", nav, errors, nil, nil}

    tmpl := template.Must(template.ParseFiles("html/index.gohtml", "html/partials/accounts.gohtml"))
    err := tmpl.ExecuteTemplate(w, "content", data)

    panicOnErr(err)
}

func deleteAccount(w http.ResponseWriter, r RequestContext) {
    id := strings.Split(r.Request.URL.Path, "/")[2]

    fmt.Printf("id: %s\n", id)

    stmt, err := db.Prepare("DELETE FROM accounts WHERE id = ?")

    panicOnErr(err)

    _, err = stmt.Exec(id)

    panicOnErr(err)
    emptyResponse(w)
}

func showAccount(w http.ResponseWriter, r RequestContext) {
    id := strings.Split(r.Request.URL.Path, "/")[2]

    fmt.Printf("id: %s\n", id)

    stmt, err := db.Prepare("DELETE FROM accounts WHERE id = ?")

    panicOnErr(err)

    _, err = stmt.Exec(id)

    panicOnErr(err)
    emptyResponse(w)
}

func getNav(path string) Nav {
    return Nav {
	Items: []*NavItem {
	    {"Home", "/", path == "/"},
	    {"Transactions", "/transactions", path == "/transactions"},
	    {"Accounts", "/accounts", path == "/accounts"},
	},
    }
}

func postStore(w http.ResponseWriter, r RequestContext) {

    r.Request.ParseForm()
    
    tmpl := template.Must(template.ParseFiles("html/partials/create-transaction.gohtml"))
    
    required := []string{"amount", "date", "description"}
    errors := make(map[string]string)
    for _, key := range required {
	if !r.Request.Form.Has(key) || r.Request.Form.Get(key) == "" {
	    errors[key] = fmt.Sprintf("The %s field is missing", key) 
	}
    }

    if len(errors) > 0 {
	// If has errors return form with errors
	nav := getNav(r.Request.URL.Path)
	data := PageData{"Page", nav, errors, nil, nil}
	err := tmpl.ExecuteTemplate(w, "content", data)

	panicOnErr(err)
	return
    }

    // otherwise continue inserting
    amount := r.Request.Form.Get("amount")
    date := r.Request.Form.Get("date")
    description := r.Request.Form.Get("description")


    stmt, err := db.Prepare("insert into transactions (account_id, amount, date, description) VALUES (?, ?, ?, ?)")

    panicOnErr(err)
    accountRepo := database.MakeAccountRepo(db)
    account := accountRepo.GetFirstAccount()

    _, err = stmt.Exec(account.Id, amount, date, description)
    defer stmt.Close()

    if err != nil {
	errServer(w, err)
	return
    }

    w.Header().Add("HX-Trigger", "new-transactions")
    err = tmpl.ExecuteTemplate(w, "content", nil)
    panicOnErr(err)
}

func truncate(w http.ResponseWriter, r RequestContext) {
    _, err := db.Exec("DELETE FROM transactions")
    panicOnErr(err)
    w.Header().Add("HX-Trigger", "new-transactions")
    serveResponse(w, "ok")
}

func postImport(w http.ResponseWriter, r RequestContext) {

    err := r.Request.ParseMultipartForm(maxMemoryFormInBytes)
    if err != nil {
	errBadRequest(w, err.Error())
	return
    }

    file, fileHeader, err := r.Request.FormFile("file")
    if err != nil {
	errBadRequest(w, err.Error())
	return
    }

    bytes := make([]byte, fileHeader.Size)

    _, err = file.Read(bytes)
    fileContent := string(bytes[:])

    // transactions, err := ReadTransactions(fileContent)
    transactions, err := ReadFromCSV(fileContent)
    panicOnErr(err)
    repo := database.MakeAccountRepo(db)
    account := repo.GetFirstAccount()
	//
    for _, transaction := range transactions {
	// Check if exists
	stmt, err := db.Prepare(`
	    SELECT COUNT(*) FROM transactions
	    WHERE external_transaction_id = ?
	`)
	panicOnErr(err)

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
                external_transaction_id,
		account_id
	    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	panicOnErr(err)

	_, err = stmt.Exec(
	    transaction.Amount,
	    date,
	    transaction.Memo,
	    transaction.Payee,
	    transaction.Address,
	    transaction.Category,
	    transaction.TransctionId,
	    account.Id,
	)
	panicOnErr(err)

	err = stmt.Close()
	panicOnErr(err)
    }
    tmpl := template.Must(template.ParseFiles("html/partials/create-transaction.gohtml"))

    // If has errors return form with errors
    nav := getNav(r.Request.URL.Path)
    messages := make(map[string]string)
    messages["import"] = "Sucessfully imported";

    w.Header().Add("HX-Trigger", "new-transactions")

    data := PageData{"Page", nav, nil, messages, nil}
    err = tmpl.ExecuteTemplate(w, "content", data)

    panicOnErr(err)
}

func deleteTransaction(w http.ResponseWriter, r RequestContext) {
    id := strings.Split(r.Request.URL.Path, "/")[2]

    fmt.Printf("id: %s\n", id)

    stmt, err := db.Prepare("DELETE FROM transactions WHERE id = ?")

    panicOnErr(err)

    _, err = stmt.Exec(id)

    panicOnErr(err)
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

func emptyResponse(
    w http.ResponseWriter,
) {
    w.Header().Add("Content-type", "text/html")
    w.WriteHeader(200)
}

func errServer(w http.ResponseWriter, err error) {
    w.WriteHeader(500);
    formattedError := fmt.Sprintf("<div>%s<div>", err.Error())

    w.Write([]byte(formattedError))
}

func errMethodNotAllowed(w http.ResponseWriter) {
    w.WriteHeader(405)
}

func errBadRequest(w http.ResponseWriter, error string) {
    w.WriteHeader(400)
    serveResponse(w, error)
}
