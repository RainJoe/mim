package model

// ImMessageSend is a mapping object for im_message_send table in postgresql
type ImMessageSend struct {
	MsgID int64 `json:"msg_id" db:"msg_id"`
	MsgFrom string `json:"msg_from" db:"msg_from"`
	MsgTo string `json:"msg_to" db:"msg_to"`
	MsgSeq int64 `json:"msg_seq" db:"msg_seq"`
	MsgContent string `json:"msg_content" db:"msg_content"`
	SendTime int64 `json:"send_time" db:"send_time"`
	MsgType int `json:"msg_type" db:"msg_type"`
}
