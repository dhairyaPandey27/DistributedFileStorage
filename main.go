package main

import (
	"bytes"
	"log"
	"strings"
	"time"

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

		StorageRoot: strings.TrimPrefix(listenAddr,":")+"_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport: tcpTransport,
		BootstrapNodes:  nodes,
	}

	
	s := NewFileServer(fileServerOpts)

	tcpTransport.OnPeer=s.onPeer

	return s

}

func main() {

	s1:=makeServer(":3000","")
	s2:=makeServer(":4000",":3000")

	go func(){
		log.Fatal(s1.Start())
	}()



	time.Sleep(1*time.Second)
	go s2.Start()
	time.Sleep(1*time.Second)

	data := bytes.NewReader([]byte("my big data file here!!"))

	s2.StoreData("myprivatedata",data)

	select{}

}