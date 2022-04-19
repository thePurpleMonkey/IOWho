package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Transaction is a struct that models the structure of a transaction, both in the request body, and in the DB
type Transaction struct {
	TransactionID int64      `json:"transaction_id"`
	Description   string     `json:"description"`
	Amount        Currency   `json:"amount"`
	Timestamp     *time.Time `json:"timestamp"`
	Notes         string     `json:"notes"`
	Contact       Contact    `json:"contact"`
}

// TransactionsHandler handles GETting all transactions and POSTing a new transaction
func TransactionsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Transactions GET - Unable to get session: %v\n", err)
		SendError(w, SERVER_ERROR_MESSAGE, http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		recent := 0

		// Check for recent transaction URL parameter
		parameters, ok := r.URL.Query()["recent"]
		if ok && len(parameters) >= 1 {
			parameter, err := strconv.Atoi(parameters[0])
			if err == nil {
				recent = parameter
			}
		}

		// Retrieve transactions in collection
		var rows *sql.Rows
		if recent > 0 {
			rows, err = db.Query("SELECT * FROM transactions_get_recent($1, $2)", session.Values["user_id"], recent)
		} else {
			rows, err = db.Query("SELECT * FROM transactions_get_all($1)", session.Values["user_id"])
		}

		if err != nil {
			log.Printf("Transactions GET - Unable to get transactions from database: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Retrieve rows from database
		transactions := make([]Transaction, 0)
		for rows.Next() {
			var transaction Transaction
			if err := rows.Scan(&transaction.TransactionID, &transaction.Description, &transaction.Amount, &transaction.Timestamp, &transaction.Notes); err != nil {
				log.Printf("Transactions GET - Unable to get transaction from database result: %v\n", err)
			}
			transactions = append(transactions, transaction)
		}

		// Check for errors from iterating over rows.
		if err := rows.Err(); err != nil {
			log.Printf("Transactions GET - Unable to get transactions from database result: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}

		// Send response
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(transactions)
		return

	} else if r.Method == "POST" {
		// Parse and decode the request body into a new `Transaction` instance
		transaction := &Transaction{}
		if err := json.NewDecoder(r.Body).Decode(transaction); err != nil {
			// If there is something wrong with the request body, return a 400 status
			log.Printf("Transactions POST - Unable to decode request body: %v\n", err)
			SendError(w, REQUEST_ERROR_MESSAGE, http.StatusBadRequest)
			return
		}

		// Input validation
		if transaction.Description == "" {
			log.Println("Transactions POST - Cannot create a transaction with a blank name.")
			SendError(w, `{"error": "Cannot add a transaction with a blank description."}`, http.StatusBadRequest)
			return
		}

		// Create collection in database
		var transactionID int64
		if err = db.QueryRow("CALL transaction_create ($1, $2, $3, $4, $5, $6)",
			session.Values["user_id"], transaction.Description, transaction.Amount, transaction.Timestamp, transaction.Notes, transaction.Contact.ContactID).Scan(&transactionID); err != nil {
			log.Printf("Transactions POST - Unable to insert transaction record in database: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}

		// Send response
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			TransactionID int64 `json:"transaction_id"`
		}{
			transactionID,
		})
		return

	}
}

// TransactionHandler handles creating, updating, and deleting a single transaction.
func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		SendError(w, SERVER_ERROR_MESSAGE, http.StatusInternalServerError)
		log.Printf("Transaction handler - Unable to get session: %v\n", err)
		return
	}

	var transaction Transaction

	// Get transaction ID from URL
	transaction.TransactionID, err = strconv.ParseInt(mux.Vars(r)["transaction_id"], 10, 64)
	if err != nil {
		log.Printf("Transaction handler - Unable to parse transaction id from URL: %v\n", err)
		SendError(w, URL_ERROR_MESSAGE, http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		// Find the transaction in the database
		if err = db.QueryRow("CALL transaction_get($1, $2)", session.Values["user_id"], transaction.TransactionID).Scan(&transaction.Description, &transaction.Amount, &transaction.Timestamp, &transaction.Notes, &transaction.Contact.ContactID); err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
			} else {
				log.Printf("Transaction GET - Unable to get transaction from database: %v\n", err)
				SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			}
			return
		}

		// Send response
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(transaction)
		return
	} else if r.Method == "PUT" {
		err := json.NewDecoder(r.Body).Decode(&transaction)
		if err != nil {
			// If there is something wrong with the request body, return a 400 status
			log.Printf("Transaction PUT - Unable to parse request body: %v\n", err)
			SendError(w, REQUEST_ERROR_MESSAGE, http.StatusBadRequest)
			return
		}

		// Update transaction in database
		if _, err = db.Exec("CALL transaction_update($1, $2, $3, $4, $5, $6, $7)", session.Values["user_id"], transaction.TransactionID, transaction.Description, transaction.Amount, transaction.Timestamp, transaction.Notes, transaction.Contact.ContactID); err != nil {
			log.Printf("Transaction PUT - Unable to update transaction in database: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	} else if r.Method == "DELETE" {
		// Start db transaction
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Transaction DELETE - Unable to start database transaction: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
		}

		// Delete transaction
		var result sql.Result
		if result, err = tx.Exec("CALL transaction_delete($1, $2)", session.Values["user_id"], transaction.TransactionID); err != nil {
			log.Printf("Transaction DELETE - Unable to delete transaction from database: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}

		// Check if a transaction was actually deleted
		var rowsAffected int64
		if rowsAffected, err = result.RowsAffected(); err != nil {
			log.Printf("Transaction DELETE - Unable to get rows affected. Assuming everything is fine? Error: %v\n", err)
		} else if rowsAffected == 0 {
			log.Printf("Transaction DELETE - No rows were deleted from the database for transaction id %d\n", transaction.TransactionID)
			SendError(w, `{"error": "No transaction was found with that ID"}`, http.StatusNotFound)
			return
		}

		// Save changes
		if err = tx.Commit(); err != nil {
			log.Printf("Transaction DELETE - Unable to commit database transaction: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}

		log.Printf("Transaction DELETE - User %d deleted transaction %d.\n", session.Values["user_id"], transaction.TransactionID)
		w.WriteHeader(http.StatusOK)
		return
	}
}
