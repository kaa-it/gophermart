package gophermart

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/kaa-it/gophermart/internal/gophermart/storage/postgres"
	authUtils "github.com/kaa-it/gophermart/pkg/auth"
	"github.com/kaa-it/gophermart/pkg/logger"
)

func RunApplication() {
	config := NewConfig()

	log, err := logger.New(config.LogLevel)
	if err != nil {
		panic(err)
	}

	log.Info("starting application")

	if err := authUtils.InitKeys(); err != nil {
		log.Fatal("failed to initialize keys: %s", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	newStorage, err := postgres.NewStorage(&postgres.Config{
		DSN: config.DatabaseURI,
	})
	if err != nil {
		panic(err)
	}

	if err := newStorage.Initialize(context.Background()); err != nil {
		panic(err)
	}
	defer newStorage.Close()

	server := NewServer(log, newStorage, config.ServerAddress)
	accrual := NewAccrual(log, newStorage, config.AccrualAddress)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		<-c
		accrual.Shutdown()
		server.Shutdown()

		wg.Done()
	}()

	go func() {
		accrual.Run()

		wg.Done()
	}()

	server.Run()

	wg.Wait()

	log.Info("application stopped")
}
