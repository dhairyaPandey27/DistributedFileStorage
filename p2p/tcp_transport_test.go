package p2p


import (
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestTCPTransport(t *testing.T){

	tcpOpts := TCPTransportOpts{
		ListenAddr: ":3000",
		Decoder: DefaultDecoder{},
		HandshakeFunc: NOPHandshakeFunc,
	}

	tr := NewTCPTransport(tcpOpts)

	assert.Equal(t,tr.ListenAddr,tcpOpts.ListenAddr)

	// Start Server

	assert.Nil(t,tr.ListenAndAccept())

}