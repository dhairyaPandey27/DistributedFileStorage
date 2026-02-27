package main

import (
	"bytes"
	"encoding/binary"
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
	gob.Register(MessageGetFile{})
}

type FileServerOpts struct {

	EncKey []byte
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


func (s *FileServer) Stream(p *Message) error{

	peers:=[]io.Writer{}

	for _,peer := range s.peers{

		peers = append(peers, peer)

	}

	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(p)

}


func (s *FileServer) Broadcast(msg *Message) error{

	buf:=new(bytes.Buffer)
	if err:=gob.NewEncoder(buf).Encode(msg); err!=nil{
		return err
	}


	for _,peer := range s.peers{
		peer.Send([]byte{p2p.IncomingMessage})
		if err:= peer.Send(buf.Bytes());err!=nil{
			return err
		}
	}

	return nil

}

type MessageGetFile struct{

	Key string

}

func(s *FileServer) Get(key string) (io.Reader ,error){

	if s.store.Has(key){
		
		fmt.Printf("[%s] serving file (%s) from local disk\n",s.Transport.Addr(),key)
		_,r,err:= s.store.Read(key)

		return r,err

	}

	fmt.Printf("[%s] dont have file (%s) locally, fetching from network.....\n",s.Transport.Addr(),key)

	msg:= Message{
		Payload: MessageGetFile{

			Key: key,

		},
	}

	if err:=s.Broadcast(& msg); err!=nil{
		return nil,err
	}

	time.Sleep(time.Millisecond*500)

	for _,peer := range s.peers{

		// First read the file size so we can Limit the amount of bytes that we read
		// from the connection, so it will not keep hanging
		var fileSize int64
		binary.Read(peer,binary.LittleEndian,&fileSize)
		
		n,err:=s.store.writeDecrypt(s.EncKey,key,io.LimitReader(peer,fileSize))
		if err!=nil{
			return nil,err
		}
		fmt.Printf("[%s] received (%d) bytes over the network from (%s)\n",s.Transport.Addr(),n,peer.RemoteAddr())

		peer.CloseStream()

	}


	_,r,err:= s.store.Read(key)
	return r,err

}

func(s *FileServer) Store(key string,r io.Reader) error{

	// 1- Store the data to the disk
	// 2- Broadcast this file to all known peers in the network

	var(

		fileBuffer =new(bytes.Buffer)
		tee =io.TeeReader(r,fileBuffer)	

	)
	size,err:=s.store.Write(key,tee)
	if err!=nil{
		return err
	}


	msg:= Message{
		Payload: MessageStoreFile{
			Key: key,
			Size: size+16,
		},
	}

	if err:=s.Broadcast(& msg);err!=nil{
		return err
	}

	time.Sleep(5*time.Millisecond)

	peers:=[]io.Writer{}
	for _,peer:= range s.peers{
		peers= append(peers, peer)
	}

	mw:=io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})
	n,err:=copyEncrypt(s.EncKey,fileBuffer,mw)
	if err!=nil{
		return err
	}


	fmt.Printf("[%s] received and written (%d) bytes to disk\n",s.Transport.Addr(),n)

	
	return nil

	

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
				log.Println("Decoding error",err)
			}	
			
			if err:= s.handleMessage(rpc.From,&m);err!=nil{
				log.Println("Handling error",err)
			}

		case <- s.quitch:
			return

		}

	}

}

func (s *FileServer) handleMessage(from string,msg *Message) error{

	switch v:=msg.Payload.(type){
	case MessageStoreFile:
		return s.handleMessageStoreFile(from,v)
	case MessageGetFile:
		return s.handleMessageGetFile(from,v)
	}


	return nil
}


func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error{

	if !s.store.Has(msg.Key){
		return fmt.Errorf("[%s] Need to serve file (%s) but it does not exist on the disk",s.Transport.Addr(),msg.Key)
	}

	fmt.Printf("[%s] serving file (%s) over the network\n",s.Transport.Addr(),msg.Key)

	fileSize,r,err:=s.store.Read(msg.Key)
	if err!=nil{
		return err
	}

	rc,ok := r.(io.ReadCloser)
	if ok{
		defer func(rc io.ReadCloser){

			fmt.Println("closing readCloser")
			defer rc.Close()

		}(rc)
	}

	peer,ok:=s.peers[from]
	if (!ok){
		return fmt.Errorf("Peer %s not in map",from)
	}

	// First send the "incomingStream" byte to the peer and then we can send
	// the file size as an int64

	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer,binary.LittleEndian,fileSize)

	n,err:=io.Copy(peer,r)
	if err!=nil{
		return err
	}

	fmt.Printf("[%s] written (%d) bytes over the network to %s\n",s.Transport.Addr(),n,from)

	return nil

}


func (s *FileServer) handleMessageStoreFile(from string,msg MessageStoreFile) error{

	peer,ok:=s.peers[from]
	if !ok{
		return fmt.Errorf("peer (%s) could not be found in peer list",from)
	}
	
	n,err := s.store.Write(msg.Key,io.LimitReader(peer,msg.Size))
	if err!=nil{
		return err
	}

	fmt.Printf("[%s] written %d bytes to disk: ",s.Transport.Addr(),n)
	peer.CloseStream()
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