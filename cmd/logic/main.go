package main

import (
	"flag"
	"github.com/RainJoe/mim/internal/logic"
	"github.com/RainJoe/mim/internal/logic/config"
	"github.com/RainJoe/mim/internal/logic/dao"
	pb "github.com/RainJoe/mim/pb/logic"
	log "github.com/RainJoe/mim/pkg/zaplog"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"net"
)

func main() {
	flag.Parse()
	if err := config.Init(); err != nil {
		log.Fatal(err)
	}
	log.Infof("loaded config %#v", config.Conf)
	conf := config.Conf
	d := dao.New(conf)
	defer d.Close()
	if err := d.PingRedis(); err != nil {
		log.Fatal(err)
	}
	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	svc := &logic.Service{
		Conf: conf,
		Dao: d,
	}
	pb.RegisterLogicServiceServer(s, svc)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
