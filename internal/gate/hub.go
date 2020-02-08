package gate

import (
	log "github.com/RainJoe/mim/pkg/zaplog"
	"github.com/RainJoe/mim/protocol"
	"github.com/golang/protobuf/proto"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	push chan *PushMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	users map[string]*Client
}

type PushMessage struct {
	uid string
	mType uint32
	message proto.Message
}

func newHub() *Hub {
	return &Hub{
		push:  make(chan *PushMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		users: make(map[string]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.users[client.uid] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.users, client.uid)
				close(client.send)
			}
		case pm := <-h.push:
			if pm.mType ==  protocol.LogoutRequestMessage {
				if client, ok := h.users[pm.uid]; ok {
					delete(h.clients, client)
					delete(h.users, client.uid)
					close(client.send)
				}
			}
			if client, ok := h.users[pm.uid]; ok {
				p, _ := protocol.NewPacket(protocol.V1, pm.mType, pm.message)
				if err := client.sendPacket(p); err != nil {
					log.Error(err)
				}
			}
		}
	}
}
