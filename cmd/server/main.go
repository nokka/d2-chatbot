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
		chatUsername  = env.String("CHAT_USERNAME", "chat")
		chatPassword  = env.String("CHAT_PASSWORD", "CivKekEmBeft")
		tradeUsername = env.String("TRADE_USERNAME", "trade")
		tradePassword = env.String("TRADE_PASSWORD", "CivKekEmBeft")
		hcUsername    = env.String("HC_USERNAME", "hc")
		hcPassword    = env.String("HC_PASSWORD", "CivKekEmBeft")
	)

	if serverAddress == "" {
		log.Println("server address not set")
		os.Exit(0)
	}

	if chatUsername == "" {
		log.Println("chat username not set")
		os.Exit(0)
	}

	if chatPassword == "" {
		log.Println("chat password not set")
		os.Exit(0)
	}

	if tradeUsername == "" {
		log.Println("trade username not set")
		os.Exit(0)
	}

	if tradePassword == "" {
		log.Println("trade password not set")
		os.Exit(0)
	}

	if hcUsername == "" {
		log.Println("hc username not set")
		os.Exit(0)
	}

	if hcPassword == "" {
		log.Println("hc password not set")
		os.Exit(0)
	}

	// Repositories
	subscriberRepository := inmem.NewRepository()

	// Chat bot connection.
	chatBot := client.New(
		serverAddress,
		chatUsername,
		chatPassword,
		subscriberRepository,
	)

	if err := chatBot.Open(); err != nil {
		log.Println("failed to open chat connection")
		os.Exit(0)
	}

	// Trade bot connection.
	tradeBot := client.New(
		serverAddress,
		tradeUsername,
		tradePassword,
		subscriberRepository,
	)

	if err := tradeBot.Open(); err != nil {
		log.Println("failed to open trade connection")
		os.Exit(0)
	}

	// HC bot connection.
	hcBot := client.New(
		serverAddress,
		hcUsername,
		hcPassword,
		subscriberRepository,
	)

	if err := hcBot.Open(); err != nil {
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
