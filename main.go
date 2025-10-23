package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/go-sql-driver/mysql"
)

type Book struct {
	ID          int
	Title       string
	Author      string
	PublishedAt string
	ISBN        string
	Price       float64
}

var db *sql.DB
// var templates *template.Template
var templates = template.Must(template.ParseGlob("templates/*.html"))


func main() {
	var err error

	// Update your MySQL connection info here
	// Format: username:password@tcp(host:port)/dbname?parseTime=true
	dsn := "root:MyDnDb6939$@tcp(127.0.0.1:3306)/bookdb?parseTime=true"

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("MySQL connection failed: %v", err)
	}

	if err := createSchema(); err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}

	templates = template.Must(template.ParseGlob("templates/*.html"))

	r := mux.NewRouter()
	r.HandleFunc("/", ListBooks).Methods("GET")
	r.HandleFunc("/books/new", newBookFormHandler).Methods("GET")
	r.HandleFunc("/books", createBookHandler).Methods("POST")
	r.HandleFunc("/books/{id:[0-9]+}", showBookHandler).Methods("GET")
	r.HandleFunc("/books/{id:[0-9]+}/edit", editBookFormHandler).Methods("GET")
	r.HandleFunc("/books/{id:[0-9]+}", updateBookHandler).Methods("POST")
	r.HandleFunc("/books/{id:[0-9]+}/delete", deleteBookHandler).Methods("POST")

	fmt.Println("Server running at: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func createSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS books (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		author VARCHAR(255) NOT NULL,
		published_at VARCHAR(100),
		isbn VARCHAR(100),
		price DECIMAL(10,2)
	);
	`
	_, err := db.Exec(schema)
	return err
}

// --- Handlers ---

func ListBooks(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, title, author FROM books")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var books []Book
    for rows.Next() {
        var book Book
        rows.Scan(&book.ID, &book.Title, &book.Author)
        books = append(books, book)
    }

    // render template
    templates.ExecuteTemplate(w, "index.html", books)
}


func newBookFormHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Action": "/books",
		"Book":   Book{},
		"Mode":   "Create",
	}
	templates.ExecuteTemplate(w, "form.html", data)
}

func createBookHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	price := parseFloat(r.FormValue("price"))

	stmt, err := db.Prepare("INSERT INTO books(title, author, published_at, isbn, price) VALUES(?,?,?,?,?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(r.FormValue("title"), r.FormValue("author"), r.FormValue("published_at"), r.FormValue("isbn"), price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	http.Redirect(w, r, fmt.Sprintf("/books/%d", id), http.StatusSeeOther)
}

func showBookHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var b Book
	err := db.QueryRow("SELECT id, title, author, published_at, isbn, price FROM books WHERE id = ?", id).
		Scan(&b.ID, &b.Title, &b.Author, &b.PublishedAt, &b.ISBN, &b.Price)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.ExecuteTemplate(w, "show.html", b)
}




func editBookFormHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var b Book
	err := db.QueryRow("SELECT id, title, author, published_at, isbn, price FROM books WHERE id = ?", id).
		Scan(&b.ID, &b.Title, &b.Author, &b.PublishedAt, &b.ISBN, &b.Price)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Action": fmt.Sprintf("/books/%d", b.ID),
		"Book":   b,
		"Mode":   "Edit",
	}
	templates.ExecuteTemplate(w, "form.html", data)
}

func updateBookHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	price := parseFloat(r.FormValue("price"))

	stmt, err := db.Prepare("UPDATE books SET title=?, author=?, published_at=?, isbn=?, price=? WHERE id=?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(r.FormValue("title"), r.FormValue("author"), r.FormValue("published_at"), r.FormValue("isbn"), price, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/books/%s", id), http.StatusSeeOther)
}

func deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	stmt, err := db.Prepare("DELETE FROM books WHERE id=?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
