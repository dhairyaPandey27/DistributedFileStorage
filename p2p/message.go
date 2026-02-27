package p2p


const(
	
	IncomingMessage = 0x1
	IncomingStream = 0x2

)


// RPC holds any arbitrary data that is being sent over
// each transport between the nodes in the network
type RPC struct{

	// From holds the remote address of the network
	From string
	Payload [] byte
	// Stream tells us whether the Incoming Message is a Message or a Stream
	Stream bool
}