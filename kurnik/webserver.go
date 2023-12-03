package kurnik

import (
	"encoding/json"
	"net/http"

	"main/utils"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

type WebClientList map[string]*websocket.Conn

func (c *WebClientList) Remove(uuid string) {
	delete(*c, uuid)
}

type WebPayload struct {
	Command string      `json:"command"`
	Data    interface{} `json:"data"`
}

func (q *KurnikBot) StartWebServer() {
	q.WebClients = make(WebClientList)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		u1 := uuid.Must(uuid.NewV4(), nil)
		q.WebClients[u1.String()] = conn
		defer q.WebClients.Remove(u1.String())

		for {
			p := WebPayload{}

			_, b, err := conn.ReadMessage()
			if err != nil {
				return
			}

			if len(b) > 0 {
				err := json.Unmarshal(b, &p)

				if err != nil {
					utils.LogError(err)
					return
				}

				q.HandleWebSocketMessage(p, conn)
			}
		}
	})

	fs := http.FileServer(http.Dir("dashboard/build/"))
	http.Handle("/", fs)
	// TODO load port from settings.json
	http.ListenAndServe(":8080", nil)
}
