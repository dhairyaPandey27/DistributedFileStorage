package main

import (
	"fmt"
	"log"

	"github.com/dhairyaPandey27/DistributedFileStorage/p2p"
)

func main(){
	
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",
		Decoder: p2p.DefaultDecoder{},
		HandshakeFunc: p2p.NOPHandshakeFunc,
		OnPeer: func(p2p.Peer) error { return fmt.Errorf("Failed the onpeer func")},
	}

	tr:=p2p.NewTCPTransport(tcpOpts) 
	
	go func(){

		for{

			msg := <-tr.Consume()
			fmt.Printf("%+v\n",msg)

		}

	} ()

	if err:=tr.ListenAndAccept();err!=nil{
		log.Fatal(err)
	}

	select {}

}