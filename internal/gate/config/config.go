package config

import (
	"flag"
	"github.com/BurntSushi/toml"
)

var (
	Conf *Config
	cfgPath string
)

type Config struct {
	WebSocketGate WebSocketGateConfig `toml:"websocket"`
}

type WebSocketGateConfig struct {
	Addr            string
	ReadBufferSize  int
	WriteBufferSize int
	LogicServerAddr string
	PushServerAddr  string
}

func init() {
	flag.StringVar(&cfgPath, "cfg", "", "default config path")
}


//Init config
func Init() (err error) {
	_, err = toml.DecodeFile(cfgPath, &Conf)
	return
}