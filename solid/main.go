package main

import (
	"database/sql"
	"log"
)

func main() {
	db, err := sql.Open("sqlite3", "orders.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLiteRepo(db)
	if err := repo.InitSchema(); err != nil {
		log.Fatal(err)
	}

	emailOrders := NewOrderService(repo, NewEmailSender())
	err = emailOrders.ProceedOrder("Иван", []string{"apple", "banana"}, 10.5)
	if err != nil {
		log.Fatal(err)
	}

	smsOrders := NewOrderService(repo, NewSMSSender())
	err = smsOrders.ProceedOrder("Мария", []string{"orange"}, 7.0)
	if err != nil {
		log.Fatal(err)
	}
}
