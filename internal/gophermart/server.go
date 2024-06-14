package gophermart

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/kaa-it/gophermart/internal/gophermart/http/rest/user"
	"github.com/kaa-it/gophermart/internal/gophermart/storage"
	"github.com/kaa-it/gophermart/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Server struct {
	config *Config
}

func New(config *Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Run() {
	log, err := logger.New(s.config.LogLevel)
	if err != nil {
		panic(err)
	}

	log.Info("starting server")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	newStorage, err := storage.NewStorage(&storage.Config{
		DSN: s.config.DatabaseUri,
	})
	if err != nil {
		panic(err)
	}

	if err := newStorage.Initialize(context.Background()); err != nil {
		panic(err)
	}
	defer newStorage.Close()

	userHandler := user.NewHandler(newStorage, log)

	r := chi.NewRouter()

	r.Mount("/api/user", userHandler.Route())

	server := &http.Server{
		Addr:    s.config.ServerAddress,
		Handler: r,
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		<-c
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error(err.Error())
		}

		wg.Done()
	}()

	err = server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err.Error())
	}

	wg.Wait()

	log.Info("server stopped")
}
