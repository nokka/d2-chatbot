package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nokka/d2-chatbot/internal/bnetd"
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
		bnetdLog      = env.String("BNETD_LOG", "")
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/chat?parseTime=true", mysqlUser, mysqlPw, mysqlHost)
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

	// Get moderators to sync.
	mods, err := subscriberRepository.FindModerators()
	if err != nil {
		log.Println("failed to sync moderators")
		os.Exit(0)
	}

	inmemRepository.SyncModerators(mods)

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

	// Open file watcher for bnetd.log to listen for changes in subscribers online state.
	w := bnetd.NewWatcher(
		bnetdLog,
		inmemRepository,
		subscriberRepository,
	)

	// Start bnetd watcher.
	err = w.Start()
	if err != nil {
		log.Println("failed to open bnetd.log watcher", err)
		os.Exit(0)
	}

	go func() {
		for {
			printMemUsage()
			time.Sleep(60 * time.Minute)
		}
	}()

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

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
// For more info on the stats read docs at: https://golang.org/pkg/runtime/#MemStats
func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Println("-------------------------------------------")
	fmt.Printf("heap allocation = %v MB\n", bToMb(m.Alloc))
	fmt.Printf("go routines = %v\n", runtime.NumGoroutine())
	fmt.Println("-------------------------------------------")
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
