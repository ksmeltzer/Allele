package dashboard

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type Broadcaster struct {
	conns    map[*websocket.Conn]bool
	mutex    sync.Mutex
	upgrader websocket.Upgrader
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		conns: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (b *Broadcaster) Start(port string) {
	expectedToken := os.Getenv("BROADCASTER_AUTH_TOKEN")
	if expectedToken == "" {
		log.Fatalf("BROADCASTER_AUTH_TOKEN must be set in the environment variables")
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("auth_token")
		if token == "" || token != expectedToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := b.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}

		b.mutex.Lock()
		b.conns[conn] = true
		b.mutex.Unlock()

		go func() {
			defer func() {
				b.mutex.Lock()
				delete(b.conns, conn)
				b.mutex.Unlock()
				conn.Close()
			}()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		}()
	})

	log.Printf("Starting Broadcaster on localhost:%s", port)
	if err := http.ListenAndServe("127.0.0.1:"+port, nil); err != nil {
		log.Fatalf("Broadcaster server failed: %v", err)
	}
}

func (b *Broadcaster) Broadcast(payload interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for conn := range b.conns {
		if err := conn.WriteJSON(payload); err != nil {
			log.Printf("Broadcast error: %v", err)
			conn.Close()
			delete(b.conns, conn)
		}
	}
}
