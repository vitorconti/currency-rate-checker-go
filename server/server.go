package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

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

func getCurrencyRate() (*currencyRate, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, false, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()
	body, error := ioutil.ReadAll(res.Body)
	if error != nil {
		return nil, false, error
	}
	var c currencyRate
	error = json.Unmarshal(body, &c)
	if error != nil {
		return nil, false, error
	}
	select {
	case <-time.After(200 * time.Millisecond):

		return &c, true, nil

	case <-ctx.Done():
		return &c, false, nil

	}

}
func insertCurrencyCheck(currencyRate *currencyRate) error {
	db, err := databaseFactory()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	stmt, err := db.Prepare("INSERT INTO currency_check(code,codein,name,high,low,varbid,pctchange,bid,ask,timestamp,create_date) values(?,?,?,?,?,?,?,?,?,?,?) ")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(currencyRate.Usdbrl.Code, currencyRate.Usdbrl.Codein, currencyRate.Usdbrl.Name, currencyRate.Usdbrl.High, currencyRate.Usdbrl.Low, currencyRate.Usdbrl.VarBid, currencyRate.Usdbrl.PctChange, currencyRate.Usdbrl.Bid, currencyRate.Usdbrl.Ask, currencyRate.Usdbrl.Timestamp, currencyRate.Usdbrl.CreateDate)
	if err != nil {
		return err
	}
	return nil
}
func currencyHandler(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	currencyRateGetStruct, haveTimeout, err := getCurrencyRate()
	if err != nil {
		panic(err)
	}
	if haveTimeout {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestTimeout)
		w.Write([]byte("Timeout de consulta atingido, por favor tente novamente"))
		return
	}
	err = insertCurrencyCheck(currencyRateGetStruct)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	result, err := json.Marshal(currencyRateGetStruct.Usdbrl.Bid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(result)

}

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

func databaseFactory() (*sql.DB, error) {
	const filePath string = "../database/currency.db"
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, err
	}
	return db, nil
}
func databaseSeederHandler(db *sql.DB) error {
	const create string = `
	CREATE TABLE IF NOT EXISTS currency_check (
	id INTEGER NOT NULL PRIMARY KEY,
	code        VARCHAR(3) NOT NULL,
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
