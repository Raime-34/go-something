package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{db: db}
}

func (r *SQLiteRepo) InitSchema() error {
	_, err := r.db.Exec(`
    CREATE TABLE IF NOT EXISTS orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        customer TEXT NOT NULL,
        products TEXT NOT NULL,
        total REAL NOT NULL,
        status TEXT NOT NULL
    )`)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteRepo) CreateOrder(customer string, products []string, total float64) error {
	_, err := r.db.Exec(
		"INSERT INTO orders (customer, products, total, status) VALUES (?, ?, ?, ?)",
		customer, fmt.Sprintf("%v", products), total, "pending",
	)
	if err != nil {
		return err
	}

	return nil
}
