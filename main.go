package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const MaxHeight = 250

func main() {
	logrus.Info("started ")
	logger := logrus.New()
	logger.SetReportCaller(true)
	logger.SetLevel(logrus.ErrorLevel)
	var err error
	err = godotenv.Load()
	if err != nil {
		logger.Fatalf("Error getting .env, %v", err)
	}

	rp, err := NewRepository(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), logger)
	if err != nil {
		logger.Fatal(err)
	}
	mapi := NewMinterAPI("", rp, logger)

	wrks := os.Getenv("PARSING_WORKERS_AMOUNT")

	wks, errw := strconv.Atoi(wrks)
	if errw != nil {
		logger.Fatal(errw)
	}
	mapi.RunWorkers(wks)

	s := NewMyAPI("8080", rp, mapi, logger)
	ServeAPI(s)
}
