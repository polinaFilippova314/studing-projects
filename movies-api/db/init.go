package db

import (
    "database/sql"
    "fmt"

    _ "github.com/mattn/go-sqlite3"
)

func Init() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "./my.db?_foreign_keys=on")
    if err != nil {
        return nil, err
    }

    //check if no errors with connection
    if err := db.Ping(); err != nil {
        db.Close()
        return nil, err
    }

    if _, err := CreateTables(db); err != nil {
        db.Close()
        return nil, err
    }

    fmt.Println("Connected to SQLite")
    return db, nil
}