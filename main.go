package main

import (
	"database/sql"
	"github.com/GGjahoon/MySimpleBank/api"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	_ "github.com/lib/pq"
	"log"
)

const (
	driverName = "postgres"
	dbSource   = "postgresql://root:123456@localhost:5432/simple_bank?sslmode=disable"
	address    = "127.0.0.1:8080"
)

func main() {
	var err error
	conn, err := sql.Open(driverName, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	store := db.NewStore(conn)
	server := api.NewServer(store)
	err = server.Start(address)
	if err != nil {
		log.Fatal("server cannot start", err)
	}
}
