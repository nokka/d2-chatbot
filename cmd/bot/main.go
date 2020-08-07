package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nokka/d2-chatbot/pkg/env"
	"github.com/nokka/d2client"
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

	client := d2client.New()
	client.Open(serverAddress)
	defer client.Close()

	// Setup channel to read on.
	ch := make(chan []byte)

	// Setup output error channel.
	errors := make(chan error)

	client.Read(ch, errors)

	err := client.Login(botUsername, botPassword)
	if err != nil {
		log.Fatal(err)
	}

	client.Whisper("nokka", "Hello!")

	// Read the output from the chat onto a channel.
	for {
		select {
		// This case means we recieved data on the connection.
		case data := <-ch:
			fmt.Println(string(data))

		case err := <-errors:
			fmt.Println(err)
			break
		}
	}
}
