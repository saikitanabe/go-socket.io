package socketio

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/saikitanabe/go-engine.io"
	"github.com/saikitanabe/go-engine.io/parser"
)

type jsonMsg struct {
	Msg string `json:"msg"`
}

// TestPayloadDecodert reproduces bug with full chain decode using unicode.
func TestPayloadDecodert(t *testing.T) {
	// s := `42:420["chat message with ack",{"msg":"111"}]`
	s := `43:421["chat message with ack",{"msg":"Ã¤Ã¤"}]`
	// s := `43:421["chat message with ack",{"msg":"kälä"}]`

	// fmt.Println(len([]byte(`21["chat message with ack",{"msg":"Ã¤Ã¤"}`)))
	// fmt.Println("actual length", len([]byte(`21["chat message with ack",{"msg":"kälä"}]`)))

	buf := bytes.NewBuffer([]byte(s))

	payloadDecoder := parser.NewPayloadDecoder(buf)

	pkgDec, err := payloadDecoder.Next()
	if err != nil {
		t.Error(err)
	}

	defer pkgDec.Close()

	decoded := make([]byte, 1024)
	n, err := pkgDec.Read(decoded)

	if err != nil {
		t.Error(err)
	}

	fmt.Println("decoded", n, string(decoded)[0:n])

	saver := &FrameSaver{}

	saver.data = []FrameData{
		{
			Buffer: bytes.NewBuffer(decoded),
			Type:   engineio.MessageText,
		},
	}

	decoder := newDecoder(saver)

	decodeData := &[]interface{}{&jsonMsg{}}
	packet := packet{Data: decodeData}

	if err := decoder.Decode(&packet); err != nil {
		log.Println("socket Decode error", err)
		t.Error(err)
	}

	err = decoder.DecodeData(&packet)
	if err != nil {
		t.Error(err)
	}

	msg, err := debugJsonMsgData(&packet)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("jsonMsg.Msg", msg.Msg)

	if msg.Msg != "Ã¤Ã¤" {
		t.Error("Bad msg", msg.Msg)
	}

	// ret, err := s.socketHandler.onPacket(decoder, &p)

}

func TestDecode(t *testing.T) {
	// s := []byte{
	// 	0x34, 0x35, 0x3a, 0x34, 0x32, 0x30, 0x5b, 0x22, 0x63, 0x68, 0x61, 0x74, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x20,
	// 	0x77, 0x69, 0x74, 0x68, 0x20, 0x61, 0x63, 0x6b, 0x22, 0x2c,
	// 	0x7b, 0x22, 0x6d, 0x73, 0x67, 0x22, 0x3a, 0x22, 0x6b, 0xc3, 0x83, 0xc2, 0xa4, 0x6c, 0xc3, 0x83, 0xc2, 0xa4, 0x22, 0x7d,
	// }
	s := []byte{
		0x32, 0x30, 0x5b, 0x22, 0x63, 0x68, 0x61, 0x74, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x20,
		0x77, 0x69, 0x74, 0x68, 0x20, 0x61, 0x63, 0x6b, 0x22, 0x2c,
		0x7b, 0x22, 0x6d, 0x73, 0x67, 0x22, 0x3a, 0x22, 0x6b, 0xc3, 0x83, 0xc2, 0xa4, 0x6c, 0xc3, 0x83, 0xc2, 0xa4, 0x22, 0x7d, 0x5d,
	}

	fmt.Println("s", string(s))

	decodeData := &[]interface{}{&jsonMsg{}}
	packet := packet{Data: decodeData}

	saver := &FrameSaver{}

	// dat := `20["chat message with ack",{"msg":"kälä"}]`

	// s := `22["chat message with ack",{"msg":"Ã¤"}]`
	// s := `20["chat message with ack",{"msg":"kala"}]`

	// https://blog.golang.org/strings
	// for i := 0; i < len(dat); i++ {
	// 	fmt.Printf("%x ", dat[i])
	// }
	// fmt.Println("")
	// bat := bytes.NewBuffer([]byte(dat))

	bat := bytes.NewBuffer(s)
	saver.data = []FrameData{
		{
			Buffer: bat,
			Type:   engineio.MessageText,
		},
	}

	decoder := newDecoder(saver)
	err := decoder.Decode(&packet)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("decoder.message", decoder.message)

	if decoder.message != "chat message with ack" {
		t.Error("Wrong message", decoder.message)
	}

	err = decoder.DecodeData(&packet)
	if err != nil {
		t.Error(err)
	}

	msg, err := debugJsonMsgData(&packet)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("jsonMsg.msg", msg.Msg)

	// TODO should check kälä
	// if msg.Msg != "" {

	// }
}

func debugJsonMsgData(packet *packet) (*jsonMsg, error) {
	switch v := packet.Data.(type) {
	case *[]interface{}:
		for _, i := range *v {
			switch t := i.(type) {
			case *jsonMsg:
				return t, nil
			}
		}
	default:
		return nil, fmt.Errorf("Type Failed %+v", v)
	}

	return nil, fmt.Errorf("Failed to parse jsonMsg")
}
