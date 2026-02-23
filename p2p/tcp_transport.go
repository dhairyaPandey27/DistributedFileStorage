package p2p

import (
	"errors"
	"fmt"
	"net"
)

// TCPPeer represents the remote node over a established TCP Connection
type TCPPeer struct{

	// conn is the underlying connection of the peer
	conn net.Conn

	// if we dial and retrieve a conn => outbound == true
	// if we accept and retrieve a conn => outbound == false
	outbound bool

}

func NewTCPPeer(conn net.Conn,outbound bool) *TCPPeer{

	return &TCPPeer{
		conn: conn,
		outbound: outbound,
	}

}


// Close implements the Peer interface
func (p *TCPPeer) Close() error{

	return p.conn.Close()

}


type TCPTransportOpts struct{

	ListenAddr string
	HandshakeFunc HandshakeFunc
	Decoder  Decoder
	OnPeer func(Peer) error

}

// Defining a TCPTransport struct
type TCPTransport struct {
	TCPTransportOpts
	listener      net.Listener
	rpcch chan RPC

	// mu sync.RWMutex
	// peers map[net.Addr]Peer
}


// Making the use of TCPTransport struct
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport{

	return &TCPTransport{
		
		TCPTransportOpts: opts,
		rpcch: make(chan RPC),

	}
}

// Consume implements the transport interface, which will return a read only channel
// for reading the incoming messages received from another PEER in the network.
func (t *TCPTransport) Consume() <- chan RPC{
	return t.rpcch
}


func (t *TCPTransport) Close() error{

	return t.listener.Close()

}

func (t *TCPTransport) Dial(addr string) error{

	conn,err := net.Dial("tcp",addr)
	if err!=nil{
		return err
	}

	go t.handleConn(conn,true)
	return nil

}

func (t *TCPTransport) ListenAndAccept() error {

	var err error

	t.listener, err  =net.Listen("tcp",t.ListenAddr)
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
		if errors.Is(err,net.ErrClosed){
			return
		}
		if err != nil{
			fmt.Printf("TCP Accept errorL: %s\n",err)
		}
		
	fmt.Printf("new incoming connection %+v\n",conn)
	go t.handleConn(conn,false)

	}

}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn,outbound bool){

	var err error

	defer func(){

		fmt.Printf("Dropping peer connection: %s",err)
		conn.Close()

	} ()

	peer := NewTCPPeer(conn,outbound)

	if err:=t.HandshakeFunc(peer);err!=nil{

		return

	}

	if t.OnPeer !=nil{
		if err=t.OnPeer(peer);err!=nil{
			return
		}
	}

	// Read Loop
	rpc :=RPC{}
	// buf := make([]byte,2000)
	for{

		err:=t.Decoder.Decode(conn,&rpc);
		if err!=nil{
			return
		}
		// n,err := conn.Read(buf)
		// if err!=nil{
			
			
		// }

		rpc.From=conn.RemoteAddr()
		t.rpcch<-rpc

	}


}