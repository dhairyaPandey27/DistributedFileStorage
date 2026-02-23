package main

import (
	"log"
	// "time"

	"github.com/dhairyaPandey27/DistributedFileStorage/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer{
	
	tcptransportOpts := p2p.TCPTransportOpts{

		ListenAddr: listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
		// TODO OnPeer

	}

	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)



	fileServerOpts := FileServerOpts{

		StorageRoot: listenAddr+"_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport: tcpTransport,
		BootstrapNodes:  nodes,
	}

	
	return  NewFileServer(fileServerOpts)

}

func main() {

	s1:=makeServer(":3000")
	s2:=makeServer(":4000",":3000")

	go func(){
		log.Fatal(s1.Start())
	}()

	s2.Start()

}