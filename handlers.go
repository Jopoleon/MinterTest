package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/k0kubun/pp"

	"github.com/go-chi/chi"
)

const (
	TxDateLayout       = "2006-01-02T15:04:05"
	InvalidLayoutError = "invalid toDate format (required layout - 2006-01-02T15:04:05)"
)

// Результатом выполнения задания должен быть сервис в котором по АПИ можно получить:
// транзакции за период,
// транзакции по адресу,
// Сумма переводов по адресу (транзакции с типом 1) за период
// Например запрос на http://localhost:8080/api/transactions/from/Mx76add9b3f868497c42932ff0f45f709404795b4a
// Должен вернуть все транзакции отправитель которых
// Mx76add9b3f868497c42932ff0f45f709404795b4a
func (a *MyAPI) TxByFromAddress(w http.ResponseWriter, r *http.Request) {
	adr := chi.URLParam(r, "address")
	if adr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("set /{address}"))
		return
	}

	res, err := a.Repository.GetTransactionFrom(adr)
	if err != nil {
		a.Logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error"))
		return
	}
	json.NewEncoder(w).Encode(res)
}

func (a *MyAPI) TxByToAddress(w http.ResponseWriter, r *http.Request) {
	adr := chi.URLParam(r, "address")
	if adr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("set /{address}"))
		return
	}
	res, err := a.Repository.GetTransactionTo(adr)
	if err != nil {
		a.Logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error"))
		return
	}
	json.NewEncoder(w).Encode(res)
}

func (a *MyAPI) TxByPeriod(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	fromDate := params.Get("fromDate")
	toDate := params.Get("toDate")
	if fromDate == "" || toDate == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("set fromDate or toDate "))
		return
	}
	fD, err := time.Parse(TxDateLayout, fromDate)
	if err != nil {
		a.Logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(InvalidLayoutError))
		return
	}
	tD, err := time.Parse(TxDateLayout, toDate)
	if err != nil {
		a.Logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(InvalidLayoutError))
		return
	}
	res, err := a.Repository.GetTransactionByTimePeriod(fD, tD)
	if err != nil {
		a.Logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error"))
		return
	}
	json.NewEncoder(w).Encode(res)
}

func (a *MyAPI) TxTotalValueByPeriod(w http.ResponseWriter, r *http.Request) {
	adr := chi.URLParam(r, "address")
	if adr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("set /{address}"))
		return
	}
	params := r.URL.Query()
	fromDate := params.Get("fromDate")
	toDate := params.Get("toDate")
	pp.Println(fromDate, toDate)
	var fD, tD time.Time
	if fromDate != "" {
		var err error
		fD, err = time.Parse(TxDateLayout, fromDate)
		if err != nil {
			a.Logger.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(InvalidLayoutError))
			return
		}
	}
	if toDate != "" {
		var err error
		tD, err = time.Parse(TxDateLayout, toDate)
		if err != nil {
			a.Logger.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(InvalidLayoutError))
			return
		}
	}
	if fromDate != "" && toDate != "" {
		if !fD.Before(tD) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("fromDate must be before toDate"))
			return
		}
	}
	res, err := a.Repository.ValueByAddressAndPeriod(adr, fD, tD)
	if err != nil {
		a.Logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error"))
		return
	}
	json.NewEncoder(w).Encode(res)
}
