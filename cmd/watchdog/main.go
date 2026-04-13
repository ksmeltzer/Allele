package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"allele/internal/alerting"
	"allele/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: watchdog <SQLITE_DB_PATH>")
	}
	dbPath := os.Args[1]

	if err := storage.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to init SQLite db: %v", err)
	}

	botToken, _ := storage.GetPluginConfig("system", "TELEGRAM_BOT_TOKEN")
	chatID, _ := storage.GetPluginConfig("system", "TELEGRAM_CHAT_ID")

	if botToken == "" || chatID == "" {
		log.Println("Warning: TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID not set in SQLite config. Telegram alerts disabled.")
	}

	alerter := alerting.NewTelegramAlerter(botToken, chatID)

	// Startup check: Ensure allele_engine is running
	out, err := exec.Command("podman", "ps", "-a", "-q", "-f", "name=allele_engine").Output()
	if err == nil && len(out) == 0 {
		log.Fatal("Container allele_engine not found, please run installer")
	}

	out, err = exec.Command("podman", "ps", "-q", "-f", "name=allele_engine").Output()
	if err == nil && len(out) == 0 {
		log.Println("Engine not running at startup. Starting via podman start...")
		if err := exec.Command("podman", "start", "allele_engine").Run(); err != nil {
			log.Printf("Failed to start podman container: %v", err)
		}
	}

	listener, err := net.Listen("tcp", "127.0.0.1:9999")
	if err != nil {
		log.Fatalf("Watchdog failed to listen: %v", err)
	}
	log.Println("Watchdog daemon listening on 127.0.0.1:9999")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		log.Println("Engine connected")
		handleConnection(conn, alerter)
	}
}

func handleConnection(conn net.Conn, alerter *alerting.TelegramAlerter) {
	defer conn.Close()

	var mu sync.Mutex
	lastHeartbeat := time.Now()
	timeoutChan := make(chan struct{})
	doneChan := make(chan struct{})

	// Heartbeat monitor goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mu.Lock()
				elapsed := time.Since(lastHeartbeat)
				mu.Unlock()

				if elapsed > 15*time.Second {
					// Timeout occurred
					select {
					case timeoutChan <- struct{}{}:
					default:
					}
					return
				}
			case <-doneChan:
				return
			}
		}
	}()

	// Reader goroutine logic inline
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			var payload alerting.Payload
			if err := json.Unmarshal(scanner.Bytes(), &payload); err != nil {
				log.Printf("Invalid JSON received: %v", err)
				continue
			}

			mu.Lock()
			lastHeartbeat = time.Now()
			mu.Unlock()

			switch payload.Event {
			case alerting.EventBoot:
				msg := "🚀 Engine Booted"
				log.Println(msg)
				_ = alerter.SendAlert(msg)
			case alerting.EventHeartbeat:
				// just update lastHeartbeat, already done
			case alerting.EventCrash:
				msg := fmt.Sprintf("🚨 CRASH: %v", payload.Data)
				log.Println(msg)
				_ = alerter.SendAlert(msg)
			case alerting.EventShutdown:
				msg := "🛑 Engine Shutdown gracefully"
				log.Println(msg)
				_ = alerter.SendAlert(msg)
			}
		}

		// EOF or error
		if err := scanner.Err(); err != nil {
			log.Printf("Socket read error: %v", err)
		} else {
			log.Println("Socket closed by engine (EOF)")
		}

		select {
		case timeoutChan <- struct{}{}:
		default:
		}
	}()

	// Wait for timeout or EOF
	<-timeoutChan
	close(doneChan)

	log.Println("Watchdog triggered: Engine crashed or froze. Restarting podman container...")
	_ = alerter.SendAlert("🚨 WATCHDOG: Engine crashed or froze. Restarting podman container...")

	cmd := exec.Command("podman", "restart", "allele_engine")
	if err := cmd.Run(); err != nil {
		errStr := fmt.Sprintf("🚨 WATCHDOG ERROR: Failed to restart podman container: %v", err)
		log.Println(errStr)
		_ = alerter.SendAlert(errStr)
	} else {
		log.Println("Podman container restarted successfully")
	}
}
