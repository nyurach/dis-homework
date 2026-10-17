package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"education/internal/app"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New()
	if err != nil {
		return err
	}
	defer application.Close()

	log.Printf("Server starts: http://%s", application.Address)
	return application.Run(ctx)
}
