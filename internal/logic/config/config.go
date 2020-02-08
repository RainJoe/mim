package config

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/RainJoe/mim/pkg/pg"
	"github.com/RainJoe/mim/pkg/redis"
)

var (
	Conf *Config
	cfgPath string
)

type Config struct {
	LogicServer LogicServerConfig `toml:"logicServer"`
	Pg    pg.Config   `toml:"pg"`
	Redis redis.Config `toml:"redis"`
}

type LogicServerConfig struct {
	Addr            string
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