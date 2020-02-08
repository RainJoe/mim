package main

import (
	"context"
	"flag"
	pb "github.com/RainJoe/mim/pb/logic"
	pbpush "github.com/RainJoe/mim/pb/push"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RainJoe/mim/internal/gate"
	"github.com/RainJoe/mim/internal/gate/config"
	log "github.com/RainJoe/mim/pkg/zaplog"
)

func main() {
	flag.Parse()
	if err := config.Init(); err != nil {
		log.Fatal(err)
	}
	log.Infof("loaded config %#v", config.Conf)
	conf := config.Conf
	conn, err := grpc.Dial(conf.WebSocketGate.LogicServerAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewLogicServiceClient(conn)

	ws := gate.NewWebSocketGate(&conf.WebSocketGate, c)
	go ws.InitHub()
	router := http.NewServeMux()
	router.HandleFunc("/ws", ws.ServeWs)
	srv := http.Server{
		Addr:    conf.WebSocketGate.Addr,
		Handler: router,
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		lis, err := net.Listen("tcp", conf.WebSocketGate.PushServerAddr)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pbpush.RegisterPushServiceServer(s, gate.NewPushService(ws))
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error(err)
	}
}
