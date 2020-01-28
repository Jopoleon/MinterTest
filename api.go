package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/cors"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
)

type MyAPI struct {
	Port       string
	StartTime  time.Time
	router     *chi.Mux
	Repository Repository
	Logger     *logrus.Logger
	Minter     Requester
}

func NewMyAPI(port string, rp Repository, rq Requester, logger *logrus.Logger) *MyAPI {
	return &MyAPI{
		Port:       port,
		StartTime:  time.Time{},
		Repository: rp,
		Logger:     logger,
		Minter:     rq,
	}
}

//ServeAPI runs http server
func ServeAPI(api *MyAPI) {
	api.InitRouter()
	s := &http.Server{
		Addr:        "0.0.0.0:" + api.Port,
		Handler:     api.router,
		ReadTimeout: 2 * time.Minute,
	}
	//implementing graceful shutdown due to kubernetes sigterm
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)
		// sigterm signal sent from kubernetes, Kubernetes sends a SIGTERM signal which is different from SIGINT (Ctrl+Client).
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint
		// We received an interrupt signal, shut down.
		if err := s.Shutdown(context.Background()); err != nil {
			api.Logger.Errorf("HTTP server Shutdown: %+v \n", err)
		}
		//NIT: Каналы можно оставлять открытыми
		close(idleConnsClosed)
	}()
	logrus.Infof("serving api at http://%s", s.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		api.Logger.Error(err)
		close(idleConnsClosed)
	}

	<-idleConnsClosed

}

const routerPrefix = "/transactions"
const apiVersion = "/api"

func (a *MyAPI) InitRouter() {

	r := chi.NewRouter()
	//Enabling Cross Origin Resource Sharing
	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)
	r.Group(func(r chi.Router) {
		// Public status routes
		//http://localhost:8080/api/transactions/from/Mx76add9b3f868497c42932ff0f45f709404795b4a
		r.Route(apiVersion+routerPrefix, func(r chi.Router) {
			r.Get("/from/{address}", a.TxByFromAddress)
			r.Get("/to/{address}", a.TxByToAddress)
			r.Get("/period", a.TxByPeriod)
			r.Get("/total/{address}", a.TxTotalValueByPeriod)
		})
	})
	a.router = r
}
