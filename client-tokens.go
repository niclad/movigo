package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// "time"
var db *sql.DB

type Token struct {
	Token     string
	Timestamp int64
}

// Initalize and connect to the database with the given path
func initDB(dbPath string) *sql.DB {
	// Attempt to open the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Error opening database: ", err.Error())
		return nil
	}
	log.Print("Successfully connected to database")

	// Read the initQuery from the file
	initQuery, err := os.ReadFile("./init-tables.sql")
	if err != nil {
		log.Fatal("Error reading schema file: ", err.Error())
	}

	// Initalize the database with the SQL file contents
	_, err = db.Exec(string(initQuery))
	if err != nil {
		log.Fatal("Error creating table: ", err.Error())
	}
	log.Print("Successfully created fresh instance of client-access database.")

	// Check the database connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database: ", err.Error())
	}

	return db
}

// Generate a new client token
func GenerateToken() Token {
	token := make([]byte, 32)
	rand.Read(token)
	timestamp := time.Now().Unix()
	return Token{Token: fmt.Sprintf("%x", token), Timestamp: timestamp}
}

// Write the token to the client-access database
func (token Token) Write(db *sql.DB) error {
	// Begin the database transaction
	tx, err := db.Begin()
	stmt, err := tx.Prepare("INSERT INTO tokens (token, created_at) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the prepared statement
	_, err = stmt.Exec(token.Token, token.Timestamp)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}
