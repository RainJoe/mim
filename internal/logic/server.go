package logic

import (
	"context"
	"github.com/RainJoe/mim/internal/logic/config"
	"github.com/RainJoe/mim/internal/logic/dao"
	"github.com/RainJoe/mim/internal/logic/model"
	pb "github.com/RainJoe/mim/pb/logic"
	pbpush "github.com/RainJoe/mim/pb/push"
	log "github.com/RainJoe/mim/pkg/zaplog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"strings"
)

const (
	C2CMessage = iota
	C2GMessage
)
type Service struct{
	Conf *config.Config
	Dao *dao.Dao
}

func kickOut(s *Service, uid string, reason int32) {
	if s.Dao.IsUserOnline(uid) {
		addr := s.Dao.GetGateOfUser(uid)
		c, conn, err := newGRPCPushClient(addr+s.Conf.LogicServer.PushServerAddr)
		if err != nil {
			log.Error(err)
			return
		}
		defer conn.Close()
		kickOutReq := &pbpush.KickOutRequest{
			Uid: uid,
			Reason:               reason,
			Ts: time.Now().UnixNano() / 1e6,
		}
		_, err = c.KickOut(context.TODO(), kickOutReq)
		if err != nil {
			log.Error(err)
		}
		if err := s.Dao.DeleteUserFromRedis(uid); err != nil {
			log.Error(err)
		}
	}
}

func (s *Service) Auth(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	rsp := &pb.AuthResponse{
		Status: 0,
		Msg: "Success",
		Ts: time.Now().UnixNano() / 1e6,
		Seq: req.Seq,
	}
	if !s.Dao.IsUserValid(req.Uid) {
		rsp.Status = 1
		rsp.Msg = "user invalid"
		return rsp, nil
	}
	if pr, ok := peer.FromContext(ctx); ok {
		addr := strings.Split(pr.Addr.String(), ":")
		if err := s.Dao.SetUserGateInfo(req.Uid, addr[0]); err != nil {
			log.Error(err)
		}
	}
	return rsp, nil
}

func (s *Service) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	rsp := &pb.LogoutResponse{
		Ts: time.Now().UnixNano() / 1e6,
		Seq: req.Seq,
	}
	go kickOut(s, req.Uid, 0)
	return rsp, nil
}

func newGRPCPushClient(addr string) (pbpush.PushServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, err
	}
	c := pbpush.NewPushServiceClient(conn)
	return c, conn, nil
}

func (s *Service) C2CSend(ctx context.Context, req *pb.C2CSendRequest) (*pb.C2CSendResponse, error) {
	rsp := &pb.C2CSendResponse{
		Ts: time.Now().UnixNano() / 1e6,
		Seq: req.Seq,
	}
	msgID, err := s.Dao.SaveSendMessage(req.From, req.To, req.Content, req.Seq, req.Ts, C2CMessage)
	if err != nil {
		return nil, err
	}
	err = s.Dao.SaveRecvMessage(req.From, req.To, msgID)
	if err != nil {
		return nil, err
	}
	go func() {
		if s.Dao.IsUserOnline(req.To) {
			addr := s.Dao.GetGateOfUser(req.To)
			c, conn, err := newGRPCPushClient(addr + s.Conf.LogicServer.PushServerAddr)
			defer conn.Close()
			if err != nil {
				log.Error(err)
				return
			}
			c2cPushReq := &pbpush.C2CPushRequest{
				To:                   req.To,
				From:                 req.From,
				Seq:                  req.Seq+1,
				Content:              req.Content,
				MsgId:                msgID,
				Ts: time.Now().UnixNano() / 1e6,
			}
			_, err = c.C2CPush(context.TODO(), c2cPushReq)
			if err != nil {
				log.Error(err)
			}
		}
	}()
	return rsp, nil
}

func (s *Service) C2GSend(ctx context.Context, req *pb.C2GSendRequest) (*pb.C2GSendResponse, error) {
	rsp := &pb.C2GSendResponse{
		Ts: time.Now().UnixNano() / 1e6,
		Seq: req.Seq,
	}
	msgID, err := s.Dao.SaveSendMessage(req.From, req.Group, req.Content, req.Seq, req.Ts, C2CMessage)
	if err != nil {
		return nil, err
	}
	table := make(map[string][]string)
	for _, uid := range	s.Dao.GetUserFromGroup(req.Group) {
		if uid == req.From {
			continue
		}
		err = s.Dao.SaveRecvMessage(req.From, uid, msgID)
		if err != nil {
			return nil, err
		}
		if s.Dao.IsUserOnline(uid) {
			addr := s.Dao.GetGateOfUser(uid) + s.Conf.LogicServer.PushServerAddr
			table[addr] = append(table[addr], uid)
		}
	}
	go func() {
		for addr, ids := range table {
			c, conn, err := newGRPCPushClient(addr)
			if err != nil {
				log.Error(err)
				continue
			}
			for _, uid := range ids {
				c2gPush := &pbpush.C2GPushRequest{
					From:                 req.From,
					To:                   uid,
					Seq:                  req.Seq+1,
					Group:                req.Group,
					Content:              req.Content,
					MsgId:             	  msgID,
					Ts: time.Now().UnixNano() / 1e6,
				}
				_, err := c.C2GPush(context.TODO(), c2gPush)
				if err != nil {
					log.Error(err)
					continue
				}
			}
			conn.Close()
		}
	}()
	return rsp, nil
}

func (s *Service) C2SPull(ctx context.Context, req *pb.C2SPullMessageRequest) (*pb.C2SPullMessageResponse, error) {
	rsp := &pb.C2SPullMessageResponse{
		Ts: time.Now().UnixNano()/1e6,
		Seq: req.Seq,
	}
	msgs := make([]*model.ImMessageSend, 0)
	err := s.Dao.GetUnReadMessage(req.Uid, req.MsgId, req.Limit, &msgs)
	if err != nil {
		return nil, err
	}
	pullMsgs := make([]*pb.PullMsg, 0)
	for _, msg := range msgs {
		var pm pb.PullMsg
		pm.MsgId = msg.MsgID
		pm.Seq = msg.MsgSeq
		pm.Ts = time.Now().UnixNano() / 1e6
		pm.From = msg.MsgFrom
		if msg.MsgType == C2GMessage {
			pm.Group = msg.MsgTo
		}
		pm.Content = msg.MsgContent
		pm.SendTime = msg.SendTime
		pullMsgs = append(pullMsgs, &pm)
		if err := s.Dao.SetMsgRead(msg.MsgID); err != nil {
			log.Error(err)
		}
	}
	rsp.Msg = pullMsgs
	return rsp, nil
}

func (s *Service) C2CPushAck(ctx context.Context, response *pb.C2CPushResponse) (*pb.Response, error) {
	rsp := &pb.Response{
		Ts:                   time.Now().UnixNano()/1e6,
		Seq:                  response.Seq,
	}
	if err := s.Dao.SetMsgRead(response.MsgId); err != nil {
		return nil, err
	}
	return rsp, nil
}

func (s *Service) C2GPushAck(ctx context.Context, response *pb.C2GPushResponse) (*pb.Response, error) {
	rsp := &pb.Response{
		Ts:                   time.Now().UnixNano()/1e6,
		Seq:                  response.Seq,
	}
	if err := s.Dao.SetMsgRead(response.MsgId); err != nil {
		return nil, err
	}
	return rsp, nil
}
