package gate

import (
	"bytes"
	"context"
	"github.com/golang/protobuf/proto"
	"time"

	pb "github.com/RainJoe/mim/pb/logic"
	log "github.com/RainJoe/mim/pkg/zaplog"
	"github.com/RainJoe/mim/protocol"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	hub *Hub

	uid string

	logicService pb.LogicServiceClient
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("error: %v", err)
			}
			break
		}
		go c.handleRecvMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func parseMessage(message []byte) (*protocol.Packet, error) {
	r := bytes.NewReader(message)
	p := protocol.Packet{}
	if err := p.UnPack(r); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) sendMessage2Write(b []byte) {
	c.send <- b
}

func (c *Client) handleHeartBeat(p *protocol.Packet) error {
	sendPacket, err := protocol.NewPacket(protocol.V1, protocol.HeartBeatResponseMessage, nil)
	if err != nil {
		return err
	}
	b, err := sendPacket.Pack()
	if err != nil {
		return err
	}
	c.sendMessage2Write(b)
	return nil
}

func (c *Client) handleAuth(p *protocol.Packet) error {
	req := pb.AuthRequest{}
	if err := proto.Unmarshal(p.Body(), &req); err != nil {
		return err
	}
	rsp, err := c.logicService.Auth(context.TODO(), &req)
	if err != nil {
		return err
	}
	sendPacket, err := protocol.NewPacket(protocol.V1, protocol.AuthResponseMessage, rsp)
	if err != nil {
		return err
	}
	if err := c.sendPacket(sendPacket); err != nil {
		return err
	}
	if rsp.Status == 0 {
		c.uid = req.Uid
		c.hub.register <- c
	}
	return nil
}

func (c *Client) handleLogout(p *protocol.Packet) error {
	req := pb.LogoutRequest{}
	if err := proto.Unmarshal(p.Body(), &req); err != nil {
		return err
	}
	rsp, err := c.logicService.Logout(context.TODO(), &req)
	if err != nil {
		return err
	}
	sendPacket, err := protocol.NewPacket(protocol.V1, protocol.LogoutResponseMessage, rsp)
	if err != nil {
		return err
	}
	if err := c.sendPacket(sendPacket); err != nil {
		return err
	}
	return nil
}

func (c *Client) sendPacket(p *protocol.Packet) error {
	msg, err := p.Pack()
	if err != nil {
		return err
	}
	c.sendMessage2Write(msg)
	return nil
}

func (c *Client) handleC2CSendRequest(p *protocol.Packet) error {
	req := pb.C2CSendRequest{}
	if err := proto.Unmarshal(p.Body(), &req); err != nil {
		return err
	}
	rsp, err := c.logicService.C2CSend(context.TODO(), &req)
	if err != nil {
		return err
	}
	sendPacket, err := protocol.NewPacket(protocol.V1, protocol.C2CSendResponseMessage, rsp)
	if err != nil {
		return err
	}
	if err := c.sendPacket(sendPacket); err != nil {
		return err
	}
	return nil
}

func (c *Client) handleC2GSendRequest(p *protocol.Packet) error {
	req := pb.C2GSendRequest{}
	if err := proto.Unmarshal(p.Body(), &req); err != nil {
		return err
	}
	rsp, err := c.logicService.C2GSend(context.TODO(), &req)
	if err != nil {
		return err
	}
	sendPacket, err := protocol.NewPacket(protocol.V1, protocol.C2GSendResponseMessage, rsp)
	if err != nil {
		return err
	}
	if err := c.sendPacket(sendPacket); err != nil {
		return err
	}
	return nil
}

func (c *Client) handleC2SPullRequest(p *protocol.Packet) error {
	req := pb.C2SPullMessageRequest{}
	if err := proto.Unmarshal(p.Body(), &req); err != nil {
		return err
	}
	rsp, err := c.logicService.C2SPull(context.TODO(), &req)
	if err != nil {
		return err
	}
	sendPacket, err := protocol.NewPacket(protocol.V1, protocol.C2SPullResponseMessage, rsp)
	if err != nil {
		return err
	}
	if err := c.sendPacket(sendPacket); err != nil {
		return err
	}
	return nil
}

func (c *Client) handleC2CPushResponse(p *protocol.Packet) error {
	req := pb.C2CPushResponse{}
	if err := proto.Unmarshal(p.Body(), &req); err != nil {
		return err
	}
	_, err := c.logicService.C2CPushAck(context.TODO(), &req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) handleC2GPushResponse(p *protocol.Packet) error {
	req := pb.C2GPushResponse{}
	if err := proto.Unmarshal(p.Body(), &req); err != nil {
		return err
	}
	_, err := c.logicService.C2GPushAck(context.TODO(), &req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) dispatchPacket(p *protocol.Packet) error {
	switch p.Header.Cmd {
	case protocol.HeartBeatRequestMessage:
		return c.handleHeartBeat(p)
	case protocol.AuthRequestMessage:
		return c.handleAuth(p)
	case protocol.LogoutRequestMessage:
		return c.handleLogout(p)
	case protocol.LogoutResponseMessage:
		return c.handleLogout(p)
	case protocol.C2CSendRequestMessage:
		return c.handleC2CSendRequest(p)
	case protocol.C2GSendRequestMessage:
		return c.handleC2GSendRequest(p)
	case protocol.C2SPullRequestMessage:
		return c.handleC2SPullRequest(p)
	case protocol.C2CPushResponseMessage:
		return c.handleC2CPushResponse(p)
	case protocol.C2GPushResponseMessage:
		return c.handleC2GPushResponse(p)
	}
	return nil
}

func (c *Client) handleRecvMessage(message []byte) {
	p, err := parseMessage(message)
	if err != nil {
		log.Error(err)
		return
	}
	if err := c.dispatchPacket(p); err != nil {
		log.Error(err)
		return
	}
}
