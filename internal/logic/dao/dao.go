package dao

import (
	"fmt"
	"github.com/RainJoe/mim/internal/logic/config"
	"github.com/RainJoe/mim/internal/logic/model"
	"github.com/RainJoe/mim/pkg/pg"
	rs "github.com/RainJoe/mim/pkg/redis"
	log "github.com/RainJoe/mim/pkg/zaplog"
	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
)

//Dao database access object
type Dao struct {
	DB        *sqlx.DB
	redisPool *redis.Pool
}

//New dao
func New(c *config.Config) (d *Dao) {
	return &Dao{
		DB:        pg.NewPostgres(&c.Pg),
		redisPool: rs.NewRedisPool(&c.Redis),
	}
}

func (d *Dao) PingRedis() error {
	conn := d.redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("PING")
	if err != nil {
		return err
	}
	return nil
}

//Close dao
func (d *Dao) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
	if d.redisPool != nil {
		d.redisPool.Close()
	}
}

func (d *Dao) SetUserGateInfo(uid string, addr string) error {
	conn := d.redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", uid, addr)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) DeleteUserFromRedis(uid string) error {
	conn := d.redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", uid)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) IsUserOnline(uid string) bool {
	conn := d.redisPool.Get()
	defer conn.Close()
	result, err := redis.Bool(conn.Do("EXISTS", uid))
	if err != nil {
		log.Error(err)
	}
	return result
}

func (d *Dao) IsUserValid(uid string) bool {
	return true
}

func (d *Dao) GetGateOfUser(uid string) string {
	conn := d.redisPool.Get()
	defer conn.Close()
	addr, err := redis.String(conn.Do("GET", uid))
	if err != nil {
		log.Error(err)
	}
	return addr
}

func (d *Dao) GetUserFromGroup(group string) []string {
	sql := `SELECT u_id FROM im_user_group WHERE group_id = $1`
	ids := make([]string, 0)
	if err := d.DB.Select(&ids, sql, group); err != nil {
		log.Error(err)
	}
	return ids
}

func (d *Dao) SaveSendMessage(from string, to string, content string, seq int64, ts int64, msgType int) (int64, error) {
	sql := `INSERT INTO im_message_send (msg_from, msg_to, msg_seq, msg_content, send_time, msg_type) VALUES ($1, $2, $3, $4, $5, $6) RETURNING msg_id;`
	stmt, err := d.DB.Prepare(sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var msgID int64
	err = stmt.QueryRow(from, to, seq, content, ts, msgType).Scan(&msgID)
	if err != nil {
		return 0, err
	}
	return msgID, nil
}

func (d *Dao) SaveRecvMessage(from string, to string, msgID int64) error {
	sql := `INSERT INTO im_message_receive (msg_from, msg_to, msg_id) VALUES ($1, $2, $3) RETURNING msg_id;`
	stmt, err := d.DB.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(from, to, msgID)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) SetMsgRead(msgID int64) error {
	sql := `UPDATE im_message_receive SET flag = 1 WHERE msg_id = $1`
	stmt, err := d.DB.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(msgID)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) GetUnReadMessage(uid string, msgId int64, limit int32, out *[]*model.ImMessageSend) error {
	sql := `SELECT t2.* FROM im_message_receive AS t1 LEFT JOIN im_message_send t2 ON t1.msg_id=t2.msg_id WHERE t1.flag = 0 AND t1.msg_to = $1`
	if msgId != 0 {
		sql += ` AND t1.msg_id > ` + fmt.Sprintf("%d", msgId)
	}
	if limit != 0 {
		sql += ` LIMIT ` + fmt.Sprintf("%d", limit)
	}
	log.Info(sql)
	if err := d.DB.Select(out, sql, uid); err != nil {
		return err
	}
	return nil
}