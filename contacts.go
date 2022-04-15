package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

// Contact is a struct that models the structure of a contact, both in the request body, and in the DB
type Contact struct {
	ContactID int64      `json:"contact_id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Notes     string     `json:"notes"`
	AddedOn   *time.Time `json:"added_on"`
}

// ContactsHandler handles GETting all contacts or POSTing a new contact.
func ContactsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ContactsHandler - Unable to get session: %v\n", err)
		w.Header().Add("Content-Type", "application/json")
		SendError(w, SERVER_ERROR_MESSAGE, http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		// Retrieve contacts
		rows, err := db.Query("SELECT contact_id, name, email, phone, notes FROM contacts WHERE user_id = $1", session.Values["user_id"])
		if err != nil {
			log.Printf("Contacts GET - Unable to retrieve contacts from database: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Retrieve rows from database
		contacts := make([]Contact, 0)
		for rows.Next() {
			var contact Contact
			if err := rows.Scan(&contact.ContactID, &contact.Name, &contact.Email, &contact.Phone, &contact.Notes); err != nil {
				log.Printf("Contacts GET - Unable to retrieve row from database result: %v\n", err)
				continue
			}
			contacts = append(contacts, contact)
		}

		// Check for errors from iterating over rows.
		if err := rows.Err(); err != nil {
			log.Printf("Contacts GET - Unable to contacts result from database: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}

		// Send response
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(contacts)
		return

	} else if r.Method == "POST" {
		// Parse and decode the request body into a new `Contact` instance
		contact := &Contact{}
		if err := json.NewDecoder(r.Body).Decode(contact); err != nil {
			// If there is something wrong with the request body, return a 400 status
			log.Printf("Contacts POST - Unable to decode request body: %v\n", err)
			body, _ := ioutil.ReadAll(r.Body)
			log.Printf("Body: %s\n", body)
			SendError(w, REQUEST_ERROR_MESSAGE, http.StatusBadRequest)
			return
		}

		// Input validation
		if len(contact.Name) == 0 {
			log.Println("Contacts POST - Contact name not provided.")
			SendError(w, `{"error": "No contact name supplied."}`, http.StatusBadRequest)
			return
		}

		// Create collection in database
		if _, err = db.Exec("INSERT INTO contacts(name, email, phone, notes, user_id) VALUES ($1, $2, $3, $4, $5)",
			contact.Name, contact.Email, contact.Phone, contact.Notes, session.Values["user_id"]); err != nil {
			if err.(*pq.Error).Code == "23505" {
				// Contact already exists
				SendError(w, `{"error": "Contact already exists."}`, http.StatusBadRequest)
				return
			}
			log.Printf("Contacts POST - Unable to insert contact record in database: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusBadRequest)
			return
		}

		// Respond with success
		w.WriteHeader(http.StatusCreated)
	}
}

// ContactHandler handles creating, updating, or deleting a single contact.
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("ContactsHandler - Unable to get session: %v\n", err)
		w.Header().Add("Content-Type", "application/json")
		SendError(w, SERVER_ERROR_MESSAGE, http.StatusInternalServerError)
		return
	}

	var contact Contact

	// Get contact ID from URL
	contact.ContactID, err = strconv.ParseInt(mux.Vars(r)["contact_id"], 10, 64)
	if err != nil {
		log.Printf("Contact handler - Unable to parse contact id from URL: %v\n", err)
		SendError(w, URL_ERROR_MESSAGE, http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		// Find the contact in the database
		if err = db.QueryRow("SELECT contact_id, name, email, phone, notes FROM contacts WHERE user_id = $1 AND contacts.contact_id = $2", session.Values["user_id"], contact.ContactID).Scan(&contact.ContactID, &contact.Name, &contact.Email, &contact.Phone, &contact.Notes); err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Contact GET - No contact found for user %v and contact id %v\n", session.Values["user_id"], contact.ContactID)
				w.WriteHeader(http.StatusNotFound)
			} else {
				log.Printf("Contact GET - Unable to get contact from database: %v\n", err)
				SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			}
			return
		}

		// Send response
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(contact)
		return

	} else if r.Method == "PUT" {
		err := json.NewDecoder(r.Body).Decode(&contact)
		if err != nil {
			// If there is something wrong with the request body, return a 400 status
			log.Printf("Contact PUT - Unable to parse request body: %v\n", err)
			SendError(w, REQUEST_ERROR_MESSAGE, http.StatusBadRequest)
			return
		}

		// Update contact in database
		var result sql.Result
		if result, err = db.Exec("UPDATE contacts SET name = $1, phone = $2, email = $3, notes = $4 WHERE user_id = $5 AND contact_id = $6", contact.Name, contact.Phone, contact.Email, contact.Notes, session.Values["user_id"], contact.ContactID); err != nil {
			log.Printf("Contact PUT - Unable to update contact in database: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}

		// Check if update did anything
		var rows int64
		rows, err = result.RowsAffected()
		if err != nil {
			log.Printf("Contact PUT - Database update unsuccessful: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}

		if rows == 0 {
			log.Printf("Contact PUT - Contact id '%v' not found in database for user %v: %v\n", contact.ContactID, session.Values["user_id"], err)
			SendError(w, `{"error": "Contact not found."}`, http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	} else if r.Method == "DELETE" {
		// Start db transaction
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Contact DELETE - Unable to start database transaction: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
		}

		// Delete contact
		if _, err = tx.Exec("DELETE FROM contacts WHERE user_id = $1 AND contact_id = $2", session.Values["user_id"], contact.ContactID); err != nil {
			log.Printf("Contact DELETE - Unable to delete contact from database: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}

		// Save changes
		if err = tx.Commit(); err != nil {
			log.Printf("Contact DELETE - Unable to commit database transaction: %v\n", err)
			SendError(w, DATABASE_ERROR_MESSAGE, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}
}
