package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/k0kubun/pp"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"
)

//транзакции за период,
//транзакции по адресу,
//Сумма переводов по адресу (транзакции с типом 1) за период
type Repository interface {
	InsertTransactions([]Transaction) error
	GetTransactionByHash(id string) (*Transaction, error)
	GetTransactionFrom(address string) ([]Transaction, error)
	GetTransactionTo(address string) ([]Transaction, error)
	GetTransactionByTimePeriod(from, to time.Time) ([]Transaction, error)
	ValueByAddressAndPeriod(address string, from, to time.Time) (float64, error)
}

type Storage struct {
	logger *logrus.Logger
	DB     *sqlx.DB
}

func NewRepository(user, password, dbname, host, port string, logger *logrus.Logger) (Repository, error) {
	str := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sqlx.Connect("postgres", str)
	if err != nil {
		logger.Error(err, "\n", str)
		return nil, err
	}
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(5)

	//simple db migration and also check on connection usability
	err2 := createTables(db)
	if err2 != nil {
		logger.Error(err2)
		return nil, err
	}
	err3 := createIndexes(db)
	if err3 != nil {
		logger.Error(err3)
		return nil, err
	}
	return &Storage{
		logger: logger,
		DB:     db,
	}, nil
}

func (s *Storage) GetTransactionByHash(id string) (*Transaction, error) {
	res := Transaction{}
	err := s.DB.Select(&res, "SELECT hash, raw_tx, from_ad,"+
		"nonce, gas_price, type, coin, to_ad, value, payload,"+
		"service_data, gas, gas_coin, tx_data, created_at FROM transactions WHERE hash=$1;", id)
	return &res, err
}

func (s *Storage) GetTransactionFrom(address string) ([]Transaction, error) {
	res := make([]Transaction, 0)
	err := s.DB.Select(&res, "SELECT hash, raw_tx, from_ad,"+
		"nonce, gas_price, type, coin, to_ad, value, payload,"+
		"service_data, gas, gas_coin, tx_data, created_at FROM transactions WHERE from_ad=$1;", address)
	return res, err
}

func (s *Storage) GetTransactionTo(address string) ([]Transaction, error) {
	res := make([]Transaction, 0)
	err := s.DB.Select(&res, "SELECT hash, raw_tx, from_ad,"+
		"nonce, gas_price, type, coin, to_ad, value, payload,"+
		"service_data, gas, gas_coin, tx_data, created_at FROM transactions WHERE to_ad=$1;", address)
	return res, err
}

func (s *Storage) GetTransactionByTimePeriod(from, to time.Time) ([]Transaction, error) {
	res := make([]Transaction, 0)
	err := s.DB.Select(&res, "SELECT hash, raw_tx, from_ad,"+
		"nonce, gas_price, type, coin, to_ad, value, payload,"+
		"service_data, gas, gas_coin, tx_data, created_at FROM transactions WHERE tx_data between $1 and $2;", from, to)
	return res, err
}

func (s *Storage) ValueByAddressAndPeriod(address string, from, to time.Time) (float64, error) {

	var err error
	var rows *sql.Rows
	switch {
	case !from.IsZero():
		rows, err = s.DB.Query("SELECT  SUM (value)  FROM "+
			"transactions WHERE to_ad=$1 AND tx_data >= $2;", address, from)
	case !to.IsZero():
		rows, err = s.DB.Query("SELECT SUM (value)  FROM "+
			"transactions WHERE to_ad=$1 AND tx_data <= $2;", address, to)
	case !from.IsZero() && !to.IsZero():
		rows, err = s.DB.Query("SELECT  SUM (value) FROM "+
			"transactions WHERE to_ad=$1 AND tx_data between $2 and $3;", address, from, to)
	default:
		rows, err = s.DB.Query("SELECT  SUM (value) FROM "+
			"transactions WHERE to_ad=$1;", address)
	}
	if err != nil {
		return 0, err
	} else {
		var total float64
		for rows.Next() {
			rows.Scan(&total)
		}
		pp.Println(total)
		return total, err
	}

}
func (s *Storage) InsertTransactions(trs []Transaction) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	for _, t := range trs {
		_, err := tx.Exec("INSERT INTO transactions (hash, raw_tx, from_ad,"+
			"nonce, gas_price, type, coin, to_ad, value, payload,"+
			"service_data, gas, gas_coin, tx_data, created_at) VALUES "+
			"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)",
			t.Hash, t.RawTx, t.From, t.Nonce, t.GasPrice, t.Type, t.Coin,
			t.To, t.Value, t.Payload, t.ServiceData, t.Gas, t.GasCoin, t.TxTime, time.Now())

		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return errors.WithMessage(err, "and rollback Tx error: "+rollbackErr.Error())
			}
			return err
		}
	}
	return tx.Commit()
}

func createTables(DB *sqlx.DB) error {
	_, err := DB.Exec(`CREATE TABLE IF NOT exists transactions (
	id SERIAL PRIMARY KEY,
	hash text,
	raw_tx text,
	from_ad varchar,
	nonce varchar,
	gas_price bigint,
	type integer,
	coin varchar,
	to_ad varchar,
	value numeric,
	payload varchar,
	service_data varchar,
	gas varchar,
	gas_coin varchar,
	tx_data timestamptz,
	created_at timestamptz);`)
	if err != nil {
		return err
	}
	return nil

}

func createIndexes(DB *sqlx.DB) error {
	_, err := DB.Exec(`
	CREATE INDEX IF NOT exists idx_transactions_from_ad
	   ON transactions(from_ad);
	CREATE INDEX IF NOT exists idx_transactions_to_ad
	   ON transactions(to_ad);
	CREATE INDEX IF NOT exists idx_transactions_tx_data
	   ON transactions(tx_data);`)
	if err != nil {
		return err
	}
	return nil

}
