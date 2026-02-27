package main


import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
	"github.com/dhairyaPandey27/PeerVault/p2p"
)


func makeServer(listenAddr string, nodes ...string) *FileServer{
	
	tcptransportOpts := p2p.TCPTransportOpts{

		ListenAddr: listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},

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
	s2:=makeServer(":7000","")
	s3:=makeServer(":8000",":3000",":7000")
	
	go func(){
		log.Fatal(s1.Start())
	}()
	time.Sleep(500*time.Millisecond)
	go func ()  {
		log.Fatal(s2.Start())
	}()



	time.Sleep(2*time.Second)
	go s3.Start()
	time.Sleep(2*time.Second)

	
	key := "coolpicture.jpg"
	data := bytes.NewReader([]byte("my big data file here!!"))
	s3.Store(key,data)

	time.Sleep(20*time.Second)

	if err:=s3.Delete(key);err!=nil{
		log.Fatal(err)
	}

	if err:= s3.store.Delete(s3.Id,key);err!=nil{
		log.Fatal(err)
	}


	r,err:=s3.Get(key)
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