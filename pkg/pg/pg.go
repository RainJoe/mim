package pg

import (
	"fmt"
	"github.com/jmoiron/sqlx"

	log "github.com/RainJoe/mim/pkg/zaplog"
)

//Config def
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	DB           string
	MaxOpenConns int
	MaxIdleConns int
}

//NewPostgres create a sql conn
func NewPostgres(cfg *Config) *sqlx.DB {
	db, err := sqlx.Open("postgres",
		fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s",
			cfg.Host,
			cfg.Port,
			cfg.User,
			cfg.DB,
			cfg.Password))
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	return db
}
