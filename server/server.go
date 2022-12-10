package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"
)

type currencyRate struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	db, err := databaseFactory()
	if err != nil {
		panic(err)
	}
	err = databaseSeederHandler(db)
	if err != nil {
		panic(err)
	}
	db.Close()
	http.HandleFunc("/cotacao", currencyHandler)
	http.ListenAndServe(":8080", nil)
}
func databaseFactory() (*sql.DB, error) {
	const filePath string = "./database/currency.db"
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, err
	}
	return db, nil
}
func databaseSeederHandler(db *sql.DB) error {
	const create string = `
	CREATE TABLE IF NOT EXISTS activities (
	id INTEGER NOT NULL PRIMARY KEY,
	code        VARCHAR(3) NOT NULL PRIMARY KEY,
	codein      VARCHAR(3) NOT NULL,
	name        VARCHAR(31) NOT NULL,
	high        NUMERIC(6,4) NOT NULL,
	low         NUMERIC(6,4) NOT NULL,
	varBid      NUMERIC(6,4) NOT NULL,
	pctChange   NUMERIC(4,2) NOT NULL,
	bid         NUMERIC(4,2) NOT NULL,
	ask         NUMERIC(6,4) NOT NULL,
	timestamp   INTEGER  NOT NULL,
	create_date VARCHAR(19) NOT NULL
	);`
	_, err := db.Exec(create)
	if err != nil {
		return err
	}
	return nil
}
func currencyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	select {
	case <-time.After(10 * time.Microsecond):

		w.Write([]byte("Timeout de gravacao no banco atingido, por favor tente novamente"))
	case <-time.After(200 * time.Microsecond):

		w.Write([]byte("Timeout de consulta atingido, por favor tente novamente"))

	case <-ctx.Done():
		// Imprime no comand line stdout
		log.Println("Request cancelada pelo cliente")
	}
}
