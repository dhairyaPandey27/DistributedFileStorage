package p2p


import (
	"errors"
	"fmt"
	"net"
	"sync"
)


// TCPPeer represents the remote node over a established TCP Connection
type TCPPeer struct{

	// Conn is the underlying connection of the peer
	net.Conn 	
	Outbound bool
	// Wg is used to ensure that data is read from a connection
	// in an orderly manner
	wg *sync.WaitGroup

}


// NewTCPPeer is a constructor that initialises the fields in the TCPPeer struct
func NewTCPPeer(conn net.Conn,outbound bool) *TCPPeer{

	return &TCPPeer{
		Conn: conn,
		Outbound: outbound,
		wg: &sync.WaitGroup{},
	}

}


// CloseStream closes the WaitGroup on a stream after the reading is done from a channel
func (p *TCPPeer) CloseStream(){

	p.wg.Done()

}


// Send writes data to the connection
func (p *TCPPeer) Send(b []byte) error{

	_,err:=p.Conn.Write(b)
	return err
}


// TCPTransportOpts represents how the connection is established with peers
// It will have all the configurable fields that will be used to create a new TCPTransport
type TCPTransportOpts struct{

	ListenAddr string
	HandshakeFunc HandshakeFunc
	Decoder  Decoder
	OnPeer func(Peer) error

}


// TCPTransport represents how the connection is established with peers
type TCPTransport struct {

	TCPTransportOpts
	listener      net.Listener
	rpcch chan RPC

}


// NewTCPTransport is constructor that takes in the configurable fields from the
// TCPTransportOpts struct and intialises the fields in the TCPTransport struct
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport{

	return &TCPTransport{
		
		TCPTransportOpts: opts,
		rpcch: make(chan RPC,1024),

	}
}


// Addr implements the Transport interface and return the address
// of the listening server
func (t *TCPTransport) Addr() string{

	return t.ListenAddr 

}


// Consume implements the transport interface, which will return a read only channel
// for reading the incoming messages received from another PEER in the network.
func (t *TCPTransport) Consume() <- chan RPC{
	return t.rpcch
}


// Close will stop the communication between the peers
func (t *TCPTransport) Close() error{

	return t.listener.Close()

}


// Dial here takes a listening address and connects to the 
// address on the named network and pass the net.Conn object
// to handleConn() for further operations.
// If we dial and retrieve a conn => outbound == true that means that we have
// initiated the connection to the Peer
func (t *TCPTransport) Dial(addr string) error{

	conn,err := net.Dial("tcp",addr)
	if err!=nil{
		return err
	}

	go t.handleConn(conn,true)
	return nil

}


// ListenAndAccept starts the server and also passes the new 
// net.Listener object to StartAcceptLoop() for further operations.
// If we accept and retrieve a conn => outbound == false that means the we have
// accepted the connection to our network
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


// StartAcceptLoop starts the actual connection between the
// two peers or network by passing the net.Conn object to 
// handleConn() for further operations
func (t *TCPTransport) StartAcceptLoop(){

	for{

		conn,err := t.listener.Accept()
		if errors.Is(err,net.ErrClosed){
			return
		}
		if err != nil{
			fmt.Printf("TCP Accept errorL: %s\n",err)
		}
		
	go t.handleConn(conn,false)

	}

}


// handleConn() takes in a net.Conn object and outbound and
// based on it makes a new Peer and starts waiting for the
// data from connected Peer using a channel
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
	for{
		rpc :=RPC{}

		err:=t.Decoder.Decode(conn,&rpc);
		if err!=nil{
			return
		}
		rpc.From=conn.RemoteAddr().String()
		if rpc.Stream{

			peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting....\n",conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n",conn.RemoteAddr())
			continue	
		}
		t.rpcch<-rpc

	}

}