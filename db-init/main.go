package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func main() {
	// Load environment variables for admin connection
	adminUser := os.Getenv("MYSQL_ADMIN_USER")
	adminPassword := os.Getenv("MYSQL_ADMIN_PASSWORD")
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")

	// Load environment variables for new database and user
	newDBName := os.Getenv("MYSQL_NEW_DB_NAME")
	newDBUser := os.Getenv("MYSQL_NEW_DB_USER")
	newDBPassword := os.Getenv("MYSQL_NEW_DB_PASSWORD")

	// Check if all required environment variables are set
	if adminUser == "" || adminPassword == "" || dbHost == "" || dbPort == "" || newDBName == "" || newDBUser == "" || newDBPassword == "" {
		log.Fatal("Please set all required environment variables: MYSQL_ADMIN_USER, MYSQL_ADMIN_PASSWORD, MYSQL_HOST, MYSQL_PORT, MYSQL_NEW_DB_NAME, MYSQL_NEW_DB_USER, MYSQL_NEW_DB_PASSWORD")
	}

	// Connect to MySQL as admin
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", adminUser, adminPassword, dbHost, dbPort)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL as admin: %v", err)
	}
	defer db.Close()

	// Create the new database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", newDBName))
	if err != nil {
		log.Fatalf("Failed to create database %s: %v", newDBName, err)
	}
	log.Printf("Database %s created successfully", newDBName)

	// use the new database
	_, err = db.Exec(fmt.Sprintf("USE %s;", newDBName))
	if err != nil {
		log.Fatalf("Failed to use database %s: %v", newDBName, err)
	}
	log.Printf("Database %s used successfully", newDBName)

	// Create the new user and grant permissions
	_, err = db.Exec(fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'", newDBUser, newDBPassword))
	if err != nil {
		log.Fatalf("Failed to create user %s: %v", newDBUser, err)
	}
	log.Printf("User %s created successfully", newDBUser)

	// Grant all privileges on the new database to the new user
	// For local connections only
	_, err = db.Exec(fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s'", newDBUser, newDBPassword))
	if err != nil {
		log.Fatalf("Failed to grant privileges to user %s on database %s: %v", newDBUser, newDBName, err)
	}
	log.Printf("Granted all privileges on %s to user %s", newDBName, newDBUser)

	// Flush privileges to ensure that the changes take effect
	_, err = db.Exec("FLUSH PRIVILEGES")
	if err != nil {
		log.Fatalf("Failed to flush privileges: %v", err)
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
		return
	}

	// Load the schema from a file
	schemaFile := "../db/schema.sql" // Path to your schema file
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("Error reading the schema file: %v", err)
	}
	schema := string(schemaBytes)

	// Split the schema by SQL statements (using ';' as the separator)
	statements := strings.Split(schema, ";")

	// Execute each statement individually within the transaction
	for _, statement := range statements {
		trimmedStatement := strings.TrimSpace(statement)
		if trimmedStatement == "" {
			continue
		}

		_, err = tx.Exec(trimmedStatement)
		if err != nil {
			// Rollback the transaction if an error occurs
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				log.Printf("Error executing statement: %s\nError: %v\nFailed to rollback: %v", trimmedStatement, err, rollbackErr)
				return
			}
			log.Printf("Error executing statement: %s\nError: %v. Transaction rolled back.", trimmedStatement, err)
			return
		} else {
			log.Printf("Successfully executed statement: %s", trimmedStatement)
		}
	}

	// Commit the transaction if all statements were executed successfully
	if err = tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
		return
	}

	log.Println("Schema applied successfully with transaction.")
}
