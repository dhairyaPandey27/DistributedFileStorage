package main

import (
	// "bytes"
	"bytes"
	"fmt"
	"io"
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
		
		EncKey: newEncryptionKey(),
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



	time.Sleep(2*time.Second)
	go s2.Start()
	time.Sleep(2*time.Second)

	
	key := "coolpicture.jpg"
	data := bytes.NewReader([]byte("my big data file here!!"))
	s2.Store(key,data)

	if err:= s2.store.Delete(key);err!=nil{
		log.Fatal(err)
	}

	r,err:=s2.Get(key)
	if err!=nil{
		log.Fatal(err)
	}


	b,err:=io.ReadAll(r)
	if err!=nil{
		log.Fatal(err)
	}
	fmt.Println(string(b))

	select{}

}