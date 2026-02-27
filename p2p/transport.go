package p2p


import "net"


// Peer is an interface that represents the remote node
type Peer interface{

	net.Conn
	// Send is used to write a ([] byte) to a connection 
	Send([]byte) error
	CloseStream()

}


// Transport is anything that handles the communication
// between the nodes in the network. This can be of the
// form (TCP , UDP , websockets, ....)
type Transport interface{

	// Addr provides the listening address 
	Addr() string
	// Dial helps in connecting with different peers
	Dial(string) error
	// ListenAndAccept starts the communication 
	ListenAndAccept() error
	// Consume helps in reading messages from different peers using a channel
	Consume() <- chan RPC
	// Close will stop the Communication 
	Close() error

}