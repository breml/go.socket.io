package engineio_v4_parser

import (
	"bytes"
	"errors"

	engineio_v4 "github.com/maldikhan/go.socket.io/engine.io/v4"
)

// recordSeparator delimits the packets of a batched payload, as delivered by
// the engine.io polling transport in a single HTTP response body. See the
// engine.io reference implementation's parser.
const recordSeparator = 0x1E

type EngineIOV4Parser struct{}

func (p *EngineIOV4Parser) Parse(data []byte) (*engineio_v4.Message, error) {
	if len(data) == 0 {
		return nil, errors.New("empty message")
	}
	return &engineio_v4.Message{
		Type: engineio_v4.EngineIOPacket(data[0] - 0x30),
		Data: data[1:],
	}, nil
}

// ParseBatch decodes a polling response body, which may carry several packets
// separated by the record separator, and returns one message per packet. A
// body without a separator is a single packet and yields one message, so the
// WebSocket path can use this entry point unchanged.
func (p *EngineIOV4Parser) ParseBatch(data []byte) ([]*engineio_v4.Message, error) {
	if len(data) == 0 {
		return nil, errors.New("empty message")
	}
	var msgs []*engineio_v4.Message
	for _, part := range bytes.Split(data, []byte{recordSeparator}) {
		if len(part) == 0 {
			continue
		}
		msg, err := p.Parse(part)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (p *EngineIOV4Parser) Serialize(msg *engineio_v4.Message) ([]byte, error) {
	if len(msg.Data) > 64*1024*1024 { // 64 MB limit
		return nil, errors.New("message data too large")
	}
	packet := make([]byte, 1, len(msg.Data)+1)
	packet[0] = byte(msg.Type) + 0x30
	return append(packet, msg.Data...), nil
}
