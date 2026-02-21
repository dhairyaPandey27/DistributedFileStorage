package p2p

import (
	"fmt"
	"net"
	"sync"
)

// Defining a TCPTransport struct
type TCPTransport struct {
	listenAddress string
	listener      net.Listener

	mu sync.RWMutex
	peers map[net.Addr]Peer
}


// Making the use of TCPTransport struct
func NewTCPTransport(listenAddr string) *TCPTransport{

	return &TCPTransport{
		listenAddress: listenAddr,
	}
}

 
func (t *TCPTransport) ListenAndAccept() error {

	var err error

	t.listener, err  =net.Listen("tcp",t.listenAddress)
	if err!=nil{
		return err
	}
	fmt.Println("Server Started")

	go t.StartAcceptLoop()

	return nil

}

func (t *TCPTransport) StartAcceptLoop(){

	for{

		conn,err := t.listener.Accept()
		if err != nil{
			fmt.Printf("TCP Accept errorL: %s\n",err)
		}

		go t.handleConn(conn)

	}

}

func (t *TCPTransport) handleConn(conn net.Conn){
	fmt.Printf("new incoming connection %+v\n",conn)
}