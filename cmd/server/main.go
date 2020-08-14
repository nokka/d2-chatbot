package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nokka/d2-chatbot/internal/client"
	"github.com/nokka/d2-chatbot/internal/inmem"
	"github.com/nokka/d2-chatbot/internal/mysql"
	"github.com/nokka/d2-chatbot/pkg/env"
)

func main() {
	var (
		serverAddress = env.String("SERVER_ADDRESS", "")
		mysqlHost     = env.String("MYSQL_HOST", "127.0.0.1:3306")
		mysqlUser     = env.String("MYSQL_USER", "chat_user")
		mysqlPw       = env.String("MYSQL_PASSWORD", "")
		chatUsername  = env.String("CHAT_USERNAME", "chat")
		chatPassword  = env.String("CHAT_PASSWORD", "")
		tradeUsername = env.String("TRADE_USERNAME", "trade")
		tradePassword = env.String("TRADE_PASSWORD", "")
		hcUsername    = env.String("HC_USERNAME", "hc")
		hcPassword    = env.String("HC_PASSWORD", "")
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

	// Mysql connection.
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/chat", mysqlUser, mysqlPw, mysqlHost)
	pool, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("failed to open mysql connection", err)
		os.Exit(0)
	}

	// Close the pool when we're done.
	defer pool.Close()

	err = pool.Ping()
	if err != nil {
		log.Println("failed to ping mysql connection", err)
		os.Exit(0)
	}

	// Repositories
	inmemRepository := inmem.NewSubscriberRepository()
	subscriberRepository := mysql.NewSubscriberRepository(pool)

	// Chat bot connection.
	cb := client.New(
		serverAddress,
		chatUsername,
		chatPassword,
		inmemRepository,
		subscriberRepository,
	)

	// Sync the chat bot in memory store with the persistent store.
	if err := cb.Sync(); err != nil {
		log.Println("failed to sync chat data", err)
		os.Exit(0)
	}

	// Make sure the sync has run before we open for incoming traffic.
	if err := cb.Open(); err != nil {
		log.Println("failed to open chat connection", err)
		os.Exit(0)
	}

	// Trade bot connection.
	tb := client.New(
		serverAddress,
		tradeUsername,
		tradePassword,
		inmemRepository,
		subscriberRepository,
	)

	// Sync the trade bot in memory store with the persistent store.
	if err := tb.Sync(); err != nil {
		log.Println("failed to sync trade data", err)
		os.Exit(0)
	}

	// Make sure the sync has run before we open for incoming traffic.
	if err := tb.Open(); err != nil {
		log.Println("failed to open trade connection", err)
		os.Exit(0)
	}

	// HC bot connection.
	hc := client.New(
		serverAddress,
		hcUsername,
		hcPassword,
		inmemRepository,
		subscriberRepository,
	)

	// Sync the hc bot in memory store with the persistent store.
	if err := hc.Sync(); err != nil {
		log.Println("failed to sync hc data", err)
		os.Exit(0)
	}

	// Make sure the sync has run before we open for incoming traffic.
	if err := hc.Open(); err != nil {
		log.Println("failed to open hc connection", err)
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
