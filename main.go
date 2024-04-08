package main

import (
	"database/sql"
	"html/template"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	// Open the SQLite database
	var err error
	db, err = sql.Open("sqlite3", "users.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create the users table if it doesn't exist
	createUserTable := `CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE,
        password TEXT
    )`
	if _, err = db.Exec(createUserTable); err != nil {
		panic(err)
	}

	addUser := `INSERT OR IGNORE INTO users (username, password) VALUES ("user", "pass")`
	if _, err = db.Exec(addUser); err != nil {
		panic(err)
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/success", handleSuccess)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":8080", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Check if the user exists in the database
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? AND password = ?", username, password).Scan(&count)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var tmpl *template.Template
		if count > 0 {
			tmpl = template.Must(template.ParseFiles("success.html"))
		} else {
			tmpl = template.Must(template.ParseFiles("error.html"))
		}

		tmpl.Execute(w, nil)
	} else {
	handleIndex(w, r)
	}
}

func handleSuccess(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("success.html"))
	tmpl.Execute(w, nil)
}
