package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/dhairyaPandey27/DistributedFileStorage/p2p"
)

func init(){
	gob.Register(MessageStoreFile{})
}

type FileServerOpts struct {

	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes  []string
}

type FileServer struct{	
	FileServerOpts

	LockPeer sync.Mutex
	peers map[string]p2p.Peer
		
	store *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {

	storeOpts := StoreOpts{

		root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	return &FileServer{

		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch: make(chan struct{}),
		peers: make(map[string]p2p.Peer),
	}

}

type Message struct{

	Payload any

}

type MessageStoreFile struct{

	Key string
	Size int64

}

func (s *FileServer) Broadcast(p *Message) error{

	peers:=[]io.Writer{}

	for _,peer := range s.peers{

		peers = append(peers, peer)

	}

	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(p)

}

func(s *FileServer) StoreData(key string,r io.Reader) error{

	// 1- Store the data to the disk
	// 2- Broadcast this file to all known peers in the network

	buf:=new(bytes.Buffer)
	msg:= Message{
		Payload: MessageStoreFile{
			Key: key,
			Size: 15,
		},
	}


	if err:=gob.NewEncoder(buf).Encode(msg); err!=nil{
		return err
	}


	for _,peer := range s.peers{
		if err:= peer.Send(buf.Bytes());err!=nil{
			return err
		}
	}

	time.Sleep(3*time.Second)

	payload:=[]byte("Verehfy LARGE FILE")
	for _,peer:=range s.peers{
		if err:=peer.Send(payload);err!=nil{
			return err
		}
	}

	return nil

	// buf :=new(bytes.Buffer)
	// tee:=io.TeeReader(r,buf)	
	// if err:=s.store.Write(key,tee);err!=nil{
	// 	return err
	// }

	// _,err:=io.Copy(buf,r)
	// if err!=nil{
	// 	return  err
	// }

	// p:=&DataMessage{
	// 	Key: key,
	// 	Data: buf.Bytes(),
	// }

	// fmt.Println(buf.Bytes())
	
	// return s.Broadcast(&Message{
	// 	From: "todo",
	// 	Payload: p,
	// })

}


func (s *FileServer) Stop(){

	close(s.quitch)

}

func (s *FileServer) onPeer(p p2p.Peer) error{

	s.LockPeer.Lock()
	defer s.LockPeer.Unlock()
	s.peers[p.RemoteAddr().String()]=p

	if(p.(*p2p.TCPPeer).Outbound==false){

		log.Printf("Connected with remote %s",p.RemoteAddr())
	}

	return nil

}


func (s *FileServer) loop(){

	defer func(){

		log.Println("file server stopped due to user action")
		s.Transport.Close()

	}()

	for{

		select{

		case rpc:= <-s.Transport.Consume():
			// fmt.Println("recv msg")
			var m Message
			if err:= gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&m);err!=nil{
				log.Println(err)
			}	
			
			if err:= s.handleMessage(rpc.From,&m);err!=nil{
				log.Println(err)
				return
			}

			// peer,ok:=s.peers[rpc.From]
			// if !ok{
			// 	panic("peer not found in peers map")
			// }
			
			// fmt.Printf("recv %+v\n",m.Payload)
			// b:=make([]byte,1000)
			// if _,err:=peer.Read(b);err!=nil{
			// 	panic(err)
			// }
			// // panic("dd")
			// fmt.Printf("%s",string(b))


			// peer.(*p2p.TCPPeer).Wg.Done()
			// if err:=s.handleMessage(&m);err!=nil{
			// 	log.Println(err)
			// }
		case <- s.quitch:
			return

		}

	}

}

func (s *FileServer) handleMessage(from string,msg *Message) error{

	switch v:=msg.Payload.(type){
	case MessageStoreFile:
		return s.handleMessageStoreFile(from,v)
	}

	return nil
}


func (s *FileServer) handleMessageStoreFile(from string,msg MessageStoreFile) error{

	peer,ok:=s.peers[from]
	if !ok{
		return fmt.Errorf("peer (%s) could not be found in peer list",from)
	}
	
	if err := s.store.Write(msg.Key,io.LimitReader(peer,msg.Size));err!=nil{
		return err
	}

	peer.(*p2p.TCPPeer).Wg.Done()
	return nil

}

func (p *FileServer) bootstrapNetwork() error{

	for _,addr := range p.BootstrapNodes{
		if(len(addr)==0){
			continue
		}
		go func(addr string){

			if err:=p.Transport.Dial(addr);err!=nil{

				log.Println("Dial Error: ",err)

			}
		}(addr)


	}

	return nil

}

func (s *FileServer) Start() error{

	if err:=s.Transport.ListenAndAccept();err!=nil{
		return err
	}
	s.bootstrapNetwork()
	s.loop()
	return nil
}