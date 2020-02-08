package gate

import (
	"context"
	pb "github.com/RainJoe/mim/pb/push"
	"github.com/RainJoe/mim/protocol"
	"time"
)

type PushService struct {
	gate *WebSocketGate
}

func NewPushService(gate *WebSocketGate) *PushService {
	return &PushService{gate}
}

func (s* PushService) KickOut(ctx context.Context, req *pb.KickOutRequest) (*pb.KickOutResponse, error) {
	rsp := &pb.KickOutResponse{
		Ts: time.Now().Unix() / 1e6,
	}
	s.gate.hub.push <- &PushMessage{req.Uid, protocol.LogoutRequestMessage, req}
	return rsp, nil

}
func(s* PushService) C2CPush(ctx context.Context, req *pb.C2CPushRequest) (*pb.Response, error) {
	rsp := &pb.Response{
		Ts: time.Now().Unix() / 1e6,
	}
	s.gate.hub.push <- &PushMessage{req.To,protocol.C2CPushRequestMessage, req}
	return rsp, nil
}

func(s* PushService) C2GPush(ctx context.Context, req *pb.C2GPushRequest) (*pb.Response, error) {
	rsp := &pb.Response{
		Ts: time.Now().Unix() / 1e6,
	}
	s.gate.hub.push <- &PushMessage{req.To,protocol.C2GPushRequestMessage, req}
	return rsp, nil
}