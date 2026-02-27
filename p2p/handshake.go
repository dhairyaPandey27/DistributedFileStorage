package p2p


// HandshakeFunc is a function type that defines the handshake protocol for establishing a connection with a peer.
type HandshakeFunc func(Peer) error


// NOPHandshakeFunc is a no-operation handshake function that always returns nil.
// It can be used when no handshake processing is required.
func NOPHandshakeFunc(Peer) error {
	return nil
}