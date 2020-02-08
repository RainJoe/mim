package gate

import (
	pb "github.com/RainJoe/mim/pb/logic"
	"net/http"

	"github.com/RainJoe/mim/internal/gate/config"
	log "github.com/RainJoe/mim/pkg/zaplog"
	"github.com/gorilla/websocket"
)

type WebSocketGate struct {
	upgrade *websocket.Upgrader
	logicService pb.LogicServiceClient
	hub *Hub
}

func NewWebSocketGate(conf *config.WebSocketGateConfig, logicService pb.LogicServiceClient) *WebSocketGate {
	return &WebSocketGate{
		upgrade: &websocket.Upgrader{
			ReadBufferSize:  conf.ReadBufferSize,
			WriteBufferSize: conf.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		logicService: logicService,
		hub: newHub(),
	}
}

func (ws *WebSocketGate) InitHub() {
	ws.hub.run()
}

// serveWs handles websocket requests from the peer.
func (ws *WebSocketGate) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Upgrade: ", err)
	}
	client := &Client{conn: conn, send: make(chan []byte, 256), logicService: ws.logicService, hub: ws.hub}
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
