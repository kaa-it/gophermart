package main

import (
	"github.com/joho/godotenv"
	"github.com/kaa-it/gophermart/internal/gophermart"
)

func main() {
	_ = godotenv.Load()

	gophermart.RunApplication()
}
