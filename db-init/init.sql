package main

import (
    "fmt"
    "log"
    "os"

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

    // Create the new user and grant permissions
    _, err = db.Exec(fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'", newDBUser, newDBPassword))
    if err != nil {
        log.Fatalf("Failed to create user %s: %v", newDBUser, err)
    }
    log.Printf("User %s created successfully", newDBUser)

    // Grant all privileges on the new database to the new user
    _, err = db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%'", newDBName, newDBUser))
    if err != nil {
        log.Fatalf("Failed to grant privileges to user %s on database %s: %v", newDBUser, newDBName, err)
    }
    log.Printf("Granted all privileges on %s to user %s", newDBName, newDBUser)

    // Flush privileges to ensure that the changes take effect
    _, err = db.Exec("FLUSH PRIVILEGES")
    if err != nil {
        log.Fatalf("Failed to flush privileges: %v", err)
    }
    log.Println("Privileges flushed successfully")

    
}

