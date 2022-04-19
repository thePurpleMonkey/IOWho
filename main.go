package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

func preventDirectoryListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func makeRouter() *mux.Router {
	r := mux.NewRouter()

	// Users
	r.HandleFunc("/user/login", login).Methods("POST")
	r.HandleFunc("/user/logout", logout)
	r.HandleFunc("/user/register", register).Methods("POST")
	r.HandleFunc("/user/password/forgot", requestPasswordResetEmail).Methods("POST")
	r.HandleFunc("/user/password/reset", resetPassword)
	r.HandleFunc("/user/verify_email", RequireAuthentication(VerifyEmailHandler))

	r.HandleFunc("/transactions", RequireAuthentication(TransactionsHandler)).Methods("GET", "POST")
	r.HandleFunc("/transaction/{transaction_id}", RequireAuthentication(TransactionHandler)).Methods("GET", "PUT")
	r.HandleFunc("/contacts", RequireAuthentication(ContactsHandler)).Methods("GET", "POST")
	r.HandleFunc("/contact/{contact_id}", RequireAuthentication(ContactHandler)).Methods("GET", "PUT")

	// Contact Us
	// r.HandleFunc("/contact", ContactHandler).Methods("POST")

	// r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
	// 	vars := mux.Vars(r)
	// 	title := vars["title"]
	// 	page := vars["page"]

	// 	fmt.Fprintf(w, "You've requested the book: %s on page %s\n", title, page)
	// })

	// Static files
	r.HandleFunc("/{filename}.html", HTMLHandler)
	r.HandleFunc("/", HTMLHandler)
	r.PathPrefix("/").Handler(http.StripPrefix("/", preventDirectoryListing(http.FileServer(http.Dir("static")))))

	return r
}

func main() {
	log.Println()
	log.Println("==============================")
	log.Println("Server booted")

	// Check environment variables
	if os.Getenv("SESSION_KEY") == "" {
		panic("Session key environment variable not set!")
	}

	if os.Getenv("DB_USERNAME") == "" {
		panic("Database username not set!")
	}

	if os.Getenv("DB_PASSWORD") == "" {
		panic("Database password not set!")
	}

	if os.Getenv("CERT_FILE") == "" {
		panic("Certificate path not set!")
	}

	if os.Getenv("KEY_FILE") == "" {
		panic("Key path not set!")
	}

	if os.Getenv("ADMIN_EMAIL") == "" {
		panic("Administrator email address not set!")
	}

	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8000"
	}

	// Initialize router
	r := makeRouter()

	// Connect to database
	var err error
	db, err = sql.Open("postgres", "user="+os.Getenv("DB_USERNAME")+" dbname=iowho password="+os.Getenv("DB_PASSWORD")+" sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Launch server
	log.Printf("Running on port %s\n", port)
	log.Fatal(http.ListenAndServeTLS(":"+port, os.Getenv("CERT_FILE"), os.Getenv("KEY_FILE"), handlers.RecoveryHandler()(r)))
}
