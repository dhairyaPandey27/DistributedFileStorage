package p2p


// Handshak Func....??
type HandshakeFunc func(Peer) error

func NOPHandshakeFunc(Peer) error {
	return nil
}