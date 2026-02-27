package p2p

import (
	"io"
)


// Decoder is an interface for decoding RPC messages from a reader.
type Decoder interface {
	Decode(io.Reader, *RPC) error
}


// DefaultDecoder is the default implementation of the Decoder interface.
type DefaultDecoder struct{}


// Decode reads and decodes an RPC message from the provided reader.
// It checks if the message is a stream and sets the Stream flag accordingly,
// or it reads the payload data from the reader.
func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {

	peekBuf := make([]byte, 1)

	if _, err := r.Read(peekBuf); err != nil {
		return nil
	}

	// In case of stream we are not decoding what is being sent	over the network.
	// We are just setting Stream true so we can handle that in our logic

	stream := peekBuf[0] == IncomingStream
	if stream {
		msg.Stream = true
		return nil
	}

	buf := make([]byte, 1028)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil

}