package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type LoginError struct {
	Message string
}

type Note struct {
	Title     string
	Content   string
	CreatedAt string
}

var userId = -1

func main() {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", "notes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createUserRequest := `CREATE TABLE IF NOT EXISTS users ( id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE, password TEXT)`
	if _, err = db.Exec(createUserRequest); err != nil {
		log.Fatal(err)
	}

	// Insert the default user if it doesn't exist
	_, err = db.Exec(`INSERT OR IGNORE INTO users (username, password) VALUES (?, ?)`, "user", "pass")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the notes table if it doesn't exist
	createNotesRequest := `CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		title TEXT,
		content TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id)
	)`
	if _, err = db.Exec(createNotesRequest); err != nil {
		log.Fatal(err)
	}

	// Insert sample notes for the default user
	var userId int
	err = db.QueryRow("SELECT id FROM users WHERE username = ?", "user").Scan(&userId)
	if err != nil {
		log.Fatal(err)
	}

	longMessage := `Lorem ipsum dolor sit amet, consectetur adipiscing elit.
                    <strong>Pellentesque risus mi</strong>, tempus quis placerat ut, porta nec
                    nulla. Vestibulum rhoncus ac ex sit amet fringilla. Nullam gravida purus
                    diam, et dictum <a>felis venenatis</a> efficitur. Aenean ac
                    <em>eleifend lacus</em>, in mollis lectus. Donec sodales, arcu et
                    sollicitudin porttitor, tortor urna tempor ligula, id porttitor mi magna a
                    neque. Donec dui urna, vehicula et sem eget, facilisis sodales sem.
                	neque. Donec dui urna, vehicula et sem eget, facilisis sodales sem.`

	sampleNotes := []struct {
		Title   string
		Content string
	}{
		{"Welcome!", "This is your first note."},
		{"Reminder", "Don't forget to update your notes regularly."},
		{"Important", "Remember to backup your notes."},
		{"Ideas", "Brainstorm some new ideas for your project."},
		{"Meeting", "Prepare agenda for the upcoming meeting."},
		{"Hello World", longMessage},
	}

	// Check if the notes table is empty for the "user" user
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM notes WHERE user_id = ?", userId).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		// Insert sample notes only if the notes table is empty for the user
		for _, note := range sampleNotes {
			_, err = db.Exec("INSERT INTO notes (user_id, title, content) VALUES (?, ?, ?)", userId, note.Title, note.Content)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	// Serve the index.html file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("template/index.html", "template/login-form.html"))
		tmpl.ExecuteTemplate(w, "index.html", nil)
	})

	// Handle login request
	http.HandleFunc("/login", loginHandler(db))
	http.HandleFunc("/signup", signupHandler(db))

	http.HandleFunc("/notes", notesHandler(db))
	http.HandleFunc("/add-note", addNoteHandler(db))

	http.HandleFunc("/filter-notes", func(w http.ResponseWriter, r *http.Request) {
		if userId == -1 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		filterText := r.FormValue("filter-text")

		query := `SELECT title, content, created_at FROM notes WHERE (title LIKE ? OR content LIKE ?) AND user_id = ?`
		rows, err := db.Query(query, "%"+filterText+"%", "%"+filterText+"%", userId)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var notes []Note

		for rows.Next() {
			var note Note
			if err := rows.Scan(&note.Title, &note.Content, &note.CreatedAt); err != nil {
				log.Println(err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			timeDate, err := time.Parse("2006-01-02T15:04:05Z", note.CreatedAt)
			if err != nil {
				log.Println(err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			note.CreatedAt = timeDate.Format("2 January 2006")
			notes = append(notes, note)
		}

		tmpl := template.Must(template.ParseFiles("template/notes-list.html", "template/note.html"))
		tmpl.ExecuteTemplate(w, "notesList", notes)
	})

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := db.QueryRow("SELECT id FROM users WHERE username = ? AND password = ?", username, password).Scan(&userId)
		if err != nil && err != sql.ErrNoRows {
			responseWithLoginError(w, "Произошла ошибка сервера при проверке пользователя, попробуйте позже")
			log.Println(err)
			return
		} else if err != nil {
			responseWithLoginError(w, "Пользователь с ведёнными логином или парелём не найден")
			return
		}

		w.Header().Set("HX-Redirect", "/notes") // Use the HX-Redirect header to indicate the redirect URL.
		w.WriteHeader(http.StatusOK)
	}
}

func signupHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		fmt.Println("in signupHandler")

		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
		if err != nil {
			responseWithLoginError(w, "Произошла ошибка сервера при проверке пользователя, попробуйте позже")
			log.Println(err)
			return
		}

		fmt.Printf("user count %d", count)

		if count > 0 {
			responseWithLoginError(w, "Пользователь с таким именем уже существует")
			return
		}

		_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, password)
		if err != nil {
			responseWithLoginError(w, "Произошла ошибка сервера при проверке пользователя, попробуйте позже")
			log.Println(err)
			return
		}

		err = db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userId)
		if err != nil {
			responseWithLoginError(w, "Произошла ошибка сервера при проверке пользователя, попробуйте позже")
			log.Println(err)
			return
		}

		w.Header().Set("HX-Redirect", "/notes")
		w.WriteHeader(http.StatusOK)
	}
}

func responseWithLoginError(w http.ResponseWriter, errorMessage string) {
	signupError := LoginError{Message: errorMessage}
	tmpl := template.Must(template.ParseFiles("template/login-form.html"))
	tmpl.ExecuteTemplate(w, "loginForm", signupError)
}

func notesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userId == -1 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		rows, err := db.Query("SELECT title, content, created_at FROM notes WHERE user_id = ?", userId)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var notes []Note

		for rows.Next() {
			var note Note
			if err := rows.Scan(&note.Title, &note.Content, &note.CreatedAt); err != nil {
				log.Println(err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			timeDate, err := time.Parse("2006-01-02T15:04:05Z", note.CreatedAt)
			if err != nil {
				log.Println(err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			note.CreatedAt = timeDate.Format("2 January 2006")
			notes = append(notes, note)
		}

		tmpl := template.Must(template.ParseFiles("template/notes.html", "template/notes-list.html", "template/note.html"))
		tmpl.ExecuteTemplate(w, "notes.html", notes)
	}
}

func addNoteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.FormValue("title")
		content := r.FormValue("content")

		if title == "" || content == "" {
			http.Error(w, "Note title or content is empty", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("INSERT INTO notes (user_id, title, content, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)", userId, title, content)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		note := Note{
			Title:     title,
			Content:   content,
			CreatedAt: time.Now().Format("2 January 2006"),
		}

		tmpl := template.Must(template.ParseFiles("template/note.html"))
		tmpl.ExecuteTemplate(w, "note", note)
	}
}
