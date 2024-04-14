package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", "users.db")
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

	sampleNotes := []struct {
		title   string
		content string
	}{
		{"Welcome!", "This is your first note."},
		{"Reminder", "Don't forget to update your notes regularly."},
		{"Important", "Remember to backup your notes."},
		{"Ideas", "Brainstorm some new ideas for your project."},
		{"Meeting", "Prepare agenda for the upcoming meeting."},
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
			_, err = db.Exec("INSERT INTO notes (user_id, title, content) VALUES (?, ?, ?)", userId, note.title, note.content)
			if err != nil {
				log.Fatal(err)
			}
	}
	}

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	// Serve the index.html file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Handle login request
	http.HandleFunc("/login", loginHandler(db))

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Check if the user exists in the database
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? AND password = ?", username, password).Scan(&count)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if count > 0 {
			// User exists, send success response
			w.Header().Set("HX-Trigger", "loginSuccess")
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
                <div class="column is-half" id="content">
                    <h1 id="pageTitle" class="title">Notes</h1>
                    <p class="has-text-success">Login successful!</p>
                </div>
            `))
		} else {
			// User doesn't exist, send error response
			w.Header().Set("HX-Trigger", "loginError")
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
                <div class="column is-half" id="content">
                    <h1 id="pageTitle" class="title">Login</h1>
					<p class="has-text-danger">Invalid username or password.</p>
                    <div id="loginForm">
                        <form hx-post="/login" hx-target="#content" hx-swap="outerHTML">
                            <div class="field">
                                <label class="label">Username</label>
                                <div class="control">
                                    <input class="input" type="text" name="username" required>
                                </div>
                            </div>
                            <div class="field">
                                <label class="label">Password</label>
                                <div class="control">
                                    <input class="input" type="password" name="password" required>
                                </div>
                            </div>
                            <div class="field">
                                <div class="control">
                                    <button class="button is-primary" type="submit">Login</button>
                                </div>
                            </div>
                        </form>
                    </div>
                </div>
            `))
		}
	}
}
