package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	DefaultMinterURL = "https://api.minter.stakeholder.space"
	TypeOne          = 1
	TypeThree        = 3
)

type Requester interface {
	GetTransactionBlock(height int) (*TransactionBlock, error)
}

type MinterAPI struct {
	Storage          Repository
	client           *http.Client
	url              string
	logger           *logrus.Logger
	HeightChan       chan int
	FailedHeightChan chan int
	FailedHeights    []int
	resultChan       chan []Transaction
}

func NewMinterAPI(url string, db Repository, logger *logrus.Logger) *MinterAPI {
	if len(url) == 0 {
		url = DefaultMinterURL
	}
	if logger == nil {
		logger = logrus.New()
	}
	logger.SetReportCaller(true)
	return &MinterAPI{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		url:              url,
		logger:           logger,
		Storage:          db,
		resultChan:       make(chan []Transaction),
		FailedHeights:    make([]int, 0),
		HeightChan:       make(chan int),
		FailedHeightChan: make(chan int),
	}
}

func (m *MinterAPI) GetTransactionBlock(height int) (*TransactionBlock, error) {
	u, err := url.Parse(m.url + "/block" + "?height=" + strconv.Itoa(height))
	if err != nil {
		m.logger.Error(err)
		return nil, err
	}
	resp, err := m.client.Get(u.String())
	if err != nil {
		m.logger.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		m.logger.Error(err)
		return nil, err
	}

	res := &TransactionBlock{}
	err = json.Unmarshal(b, &res)
	if err != nil {
		m.logger.Debug(err, "\n response body string: ", string(b))
		return nil, err
	}
	if res.Error != nil {
		m.logger.Error(err)
		return res, res.Error
	}
	res.Result.Transactions = parseTransactions(res)
	return res, nil
}

func parseTransactions(tb *TransactionBlock) []Transaction {
	res := make([]Transaction, 0)
	for i := 0; i < len(tb.Result.Transactions); i++ {
		switch tb.Result.Transactions[i].Type {
		case TypeOne:
			v, _ := strconv.ParseFloat(tb.Result.Transactions[i].Data.Value, 64)
			n := Transaction{
				Hash:        tb.Result.Transactions[i].Hash,
				RawTx:       tb.Result.Transactions[i].RawTx,
				From:        tb.Result.Transactions[i].From,
				Nonce:       tb.Result.Transactions[i].Nonce,
				GasPrice:    tb.Result.Transactions[i].GasPrice,
				Type:        tb.Result.Transactions[i].Type,
				TxTime:      tb.Result.Time,
				Coin:        tb.Result.Transactions[i].Data.Coin,
				To:          tb.Result.Transactions[i].Data.To,
				Value:       v,
				Payload:     tb.Result.Transactions[i].Payload,
				ServiceData: tb.Result.Transactions[i].ServiceData,
				Gas:         tb.Result.Transactions[i].Gas,
				GasCoin:     tb.Result.Transactions[i].GasCoin,
			}
			res = append(res, n)
		case TypeThree:
			v, _ := strconv.ParseFloat(tb.Result.Transactions[i].Data.Value, 64)
			n := Transaction{
				Hash:        tb.Result.Transactions[i].Hash,
				RawTx:       tb.Result.Transactions[i].RawTx,
				From:        tb.Result.Transactions[i].From,
				Nonce:       tb.Result.Transactions[i].Nonce,
				GasPrice:    tb.Result.Transactions[i].GasPrice,
				Type:        tb.Result.Transactions[i].Type,
				TxTime:      tb.Result.Time,
				Coin:        tb.Result.Transactions[i].Data.Coin,
				To:          tb.Result.Transactions[i].Data.To,
				Value:       v,
				Payload:     tb.Result.Transactions[i].Payload,
				ServiceData: tb.Result.Transactions[i].ServiceData,
				Gas:         tb.Result.Transactions[i].Gas,
				GasCoin:     tb.Result.Transactions[i].GasCoin,
			}
			res = append(res, n)
		}
	}
	return res
}

func (m *MinterAPI) SimpleParser() {
	for i := 1; i < MaxHeight+1; i++ {
		fmt.Println("Current height: ", i)
		res, err := m.GetTransactionBlock(i)
		if err != nil {
			m.logger.Error("Problem in height: ", i, "; while worker doing job err: ", err)
		}
		err2 := m.Storage.InsertTransactions(res.Result.Transactions)
		if err2 != nil {
			m.logger.Error(err2)
		}
	}
}

func (m *MinterAPI) RunWorkers(n int) {
	go func() {
		m.result()
	}()
	for i := 0; i < n; i++ {
		go func() {
			for {
				m.Worker()
			}
		}()
	}
}
func (m *MinterAPI) result() {
	go func() {
		for i := 1; i < MaxHeight+1; i++ {
			m.HeightChan <- i
		}
		//retry processing failed heights
		for _, i := range m.FailedHeights {
			m.HeightChan <- i
		}
		m.logger.Info("Downloading of transaction blocks complete!")
	}()

	for n := range m.resultChan {
		err2 := m.Storage.InsertTransactions(n)
		if err2 != nil {
			m.logger.Debug("while result doing job err: ", err2)
		}
	}
}

func (m *MinterAPI) Worker() {
	height := <-m.HeightChan
	if height%10 == 0 {
		fmt.Println("Current Height: ", height)
	}
	res, err := m.GetTransactionBlock(height)
	if err != nil {
		//TODO: save failed blocks heights and make requests later
		// m.FailedHeights <- height and then workers catch it
		m.logger.Debug("Problem in height: ", height, "; while worker doing job err: ", err)
		//saving to []int slice to process failed heights later
		m.FailedHeights = append(m.FailedHeights, height)
		return
	}
	m.resultChan <- res.Result.Transactions
}
