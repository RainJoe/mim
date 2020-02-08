package protocol

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/golang/protobuf/proto"
)

const (
	V1 = iota
)

const (
	HeaderLength     = 12
)

const (
	HeartBeatRequestMessage = iota
	HeartBeatResponseMessage
	AuthRequestMessage
	AuthResponseMessage
	LogoutRequestMessage
	LogoutResponseMessage
	C2CSendRequestMessage
	C2CSendResponseMessage
	C2CPushRequestMessage
	C2CPushResponseMessage
	C2GSendRequestMessage
	C2GSendResponseMessage
	C2GPushRequestMessage
	C2GPushResponseMessage
	C2SPullRequestMessage
	C2SPullResponseMessage
)

type Header struct {
	Version   uint32
	Cmd       uint32
	BodyLen   uint32
}

func NewHeader(version uint32, cmd uint32) Header {
	return Header{
		Version:   version,
		Cmd:       cmd,
	}
}

type Packet struct {
	Header Header
	body   []byte
}

func NewPacket(version uint32, cmd uint32, pb proto.Message) (*Packet, error) {
	h := NewHeader(version, cmd)
	p := Packet{
		Header: h,
	}
	if err := p.SetBody(pb); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *Packet) Pack() ([]byte, error) {
	buf := bytes.Buffer{}
	if err := binary.Write(&buf, binary.BigEndian, &p.Header); err != nil {
		return nil, err
	}
	_, err := buf.Write(p.body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Packet) UnPack(r io.Reader) (err error) {
	buf := make([]byte, HeaderLength)
	if _, err = io.ReadFull(r, buf); err != nil {
		return err
	}
	if err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &p.Header); err != nil {
		return err
	}
	p.body = make([]byte, p.Header.BodyLen)
	if _, err = io.ReadFull(r, p.body); err != nil {
		return err
	}
	return
}

func (p *Packet) SetBody(pb proto.Message) (err error) {
	if pb == nil {
		return
	}
	data, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	p.body = data
	p.Header.BodyLen = uint32(len(data))
	return
}

func (p *Packet) Body() []byte {
	return p.body
}
