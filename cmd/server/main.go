package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nokka/d2-chatbot/internal/client"
	"github.com/nokka/d2-chatbot/internal/inmem"
	"github.com/nokka/d2-chatbot/pkg/env"
)

func main() {
	var (
		serverAddress = env.String("SERVER_ADDRESS", "")
		botUsername   = env.String("BOT_USERNAME", "")
		botPassword   = env.String("BOT_PASSWORD", "")
	)

	if serverAddress == "" {
		log.Println("server address not set")
		os.Exit(0)
	}

	if botUsername == "" {
		log.Println("bot username not set")
		os.Exit(0)
	}

	if botPassword == "" {
		log.Println("bot password not set")
		os.Exit(0)
	}

	// Repositories
	subscriberRepository := inmem.NewInmemRepository()

	hcc := client.New(
		serverAddress,
		botUsername,
		botPassword,
		subscriberRepository,
	)

	if err := hcc.Open(); err != nil {
		log.Println("failed to open hc connection")
		os.Exit(0)
	}

	// Channel to receive errors on.
	errorChannel := make(chan error)

	// Capture interupts.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errorChannel <- fmt.Errorf("got signal %s", <-c)
	}()

	// Listen for errors indefinitely.
	if err := <-errorChannel; err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
