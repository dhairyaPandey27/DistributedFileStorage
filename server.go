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
	"github.com/dhairyaPandey27/PeerVault/p2p"
)


func init(){
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
	gob.Register(MessageDeleteFile{})
}


// FileServerOpts is a struct that holds the configuration options for the FileServer.
type FileServerOpts struct {

	Id string
	EncKey []byte
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes  []string

}


// FileServer is a struct that represents a file server in the distributed file system.
// It uses FileServerOpts for configuration and has methods to handle file storage
// retrieval, and deletion, as well as peer management and message broadcasting.
type FileServer struct{	

	FileServerOpts

	LockPeer sync.Mutex
	peers map[string]p2p.Peer
		
	store *Store
	quitch chan struct{}

}


// NewFileServer creates a new instance of FileServer with the provided options.
func NewFileServer(opts FileServerOpts) *FileServer {

	storeOpts := StoreOpts{

		root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	if len(opts.Id)==0{
		opts.Id = generateID()
	}

	return &FileServer{

		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch: make(chan struct{}),
		peers: make(map[string]p2p.Peer),
	}

}


// Message is a struct that represents a message that can be broadcasted to peers in the network.
type Message struct{
	
	Payload any
	
}


// MessageStoreFile is a struct that represents a message for storing a file in the network.
type MessageStoreFile struct{
	
	Id string
	Key string
	Size int64
	
}


// MessageGetFile is a struct that represents a message for requesting a file from the network.
type MessageGetFile struct{

	Id string
	Key string

}


// MessageDeleteFile is a struct that represents a message for deleting a file from the network.
type MessageDeleteFile struct{

	Id string
	Key string

}


// Broadcast sends a message to all known peers in the network by encoding the message and writing it to each peer's connection.
func (s *FileServer) Broadcast(msg *Message) error{

	buf:=new(bytes.Buffer)
	if err:=gob.NewEncoder(buf).Encode(msg); err!=nil{
		return err
	}


	for _,peer := range s.peers{
		// First send the "incomingMessage" byte to the peer and then we can send the actual message.
		peer.Send([]byte{p2p.IncomingMessage})
		if err:= peer.Send(buf.Bytes());err!=nil{
			return err
		}
	}

	return nil

}


// Delete deletes the file associated with the key for the specified peer ID in the local 
// storage and if not found in local storage, it broadcasts a delete message to all peers
// in the network to delete the file from their storage.
func(s *FileServer) Delete(key string) error{

	if s.store.Has(s.Id,key){
		s.store.Delete(s.Id,key)
	}

	fmt.Println("File not present locally, now deleting from all over the network")
	msg := Message{
		Payload: MessageDeleteFile{

			Id: s.Id,
			Key: hashKey(key),

		},
	}

	if err:=s.Broadcast(& msg); err!=nil{
		return err
	}



	return nil

}


// Get retrieves the file associated with the key for the specified peer ID.
// If the file is present in local storage, it returns a reader for the file.
// If not, it broadcasts a get message to all peers in the network to fetch
// the file and then returns a reader for the file once it is received.
func(s *FileServer) Get(key string) (io.Reader ,error){

	if s.store.Has(s.Id,key){
		
		fmt.Printf("[%s] serving file (%s) from local disk\n",s.Transport.Addr(),key)
		_,r,err:= s.store.Read(s.Id,key)

		return r,err

	}

	fmt.Printf("[%s] dont have file (%s) locally, fetching from network.....\n",s.Transport.Addr(),key)

	msg:= Message{
		Payload: MessageGetFile{

			Id: s.Id,
			Key: hashKey(key),

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
		
		n,err:=s.store.writeDecrypt(s.Id,s.EncKey,key,io.LimitReader(peer,fileSize))
		if err!=nil{
			return nil,err
		}
		fmt.Printf("[%s] received (%d) bytes over the network from (%s)\n",s.Transport.Addr(),n,peer.RemoteAddr())

		peer.CloseStream()

	}


	_,r,err:= s.store.Read(s.Id,key)
	return r,err

}


// Store stores the file from the provided reader to the local disk and then 
// broadcasts a message to all peers in the network to store the file in their local storage as well.
func(s *FileServer) Store(key string,r io.Reader) error{

	var(

		fileBuffer =new(bytes.Buffer)
		tee =io.TeeReader(r,fileBuffer)	

	)
	size,err:=s.store.Write(s.Id,key,tee)
	if err!=nil{
		return err
	}


	msg:= Message{
		Payload: MessageStoreFile{
			Id: s.Id,
			Key: hashKey(key),
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


// Stop stops the file server by closing the quit channel and the transport connection.
func (s *FileServer) Stop(){

	close(s.quitch)

}


// onPeer is a callback function that is called when a new peer connects to the file server.
// It adds the peer to the list of known peers and logs the connection.
func (s *FileServer) onPeer(p p2p.Peer) error{

	s.LockPeer.Lock()
	defer s.LockPeer.Unlock()
	s.peers[p.RemoteAddr().String()]=p

	if(p.(*p2p.TCPPeer).Outbound==false){

		log.Printf("Connected with remote %s",p.RemoteAddr())
	}

	return nil

}


// loop is the main loop of the file server that listens for incoming messages from peers 
// and handles them accordingly, as well as checking for quit signals to stop the server.
func (s *FileServer) loop(){

	defer func(){

		log.Println("file server stopped due to user action")
		s.Transport.Close()

	}()

	for{

		select{

		case rpc:= <-s.Transport.Consume():
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


// handleMessage is a function that handles incoming messages from peers by 
// checking the type of the message payload and calling the appropriate 
// handler function for storing, retrieving, or deleting files based on the message type.
func (s *FileServer) handleMessage(from string,msg *Message) error{

	switch v:=msg.Payload.(type){
	case MessageStoreFile:
		return s.handleMessageStoreFile(from,v)
	case MessageGetFile:
		return s.handleMessageGetFile(from,v)
	case MessageDeleteFile:
		return s.handleDeleteFile(from,v)
	}


	return nil
}


// handleDeleteFile handles the incoming delete file message by checking if the file exists
// in local storage associated with the ID attached with the msg and deleting it if it does,
// or logging a message if it does not exist.
func (s *FileServer) handleDeleteFile(from string,msg MessageDeleteFile) error{

	if s.store.Has(msg.Id,msg.Key){
		return s.store.Delete(msg.Id,msg.Key)
	}

	fmt.Println("File is not present in peer")

	return nil

}


// handleMessageGetFile handles the incoming get file message by checking if the file exists in local storage
// associated with the ID attached with the msg and serving it if it does, or logging a message if it does not exist.
func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error{

	if !s.store.Has(msg.Id,msg.Key){
		return fmt.Errorf("[%s] Need to serve file (%s) but it does not exist on the disk",s.Transport.Addr(),msg.Key)
	}

	fmt.Printf("[%s] serving file (%s) over the network\n",s.Transport.Addr(),msg.Key)

	fileSize,r,err:=s.store.Read(msg.Id,msg.Key)
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


// handleMessageStoreFile handles the incoming store file message by reading the file data from the peer's connection.
func (s *FileServer) handleMessageStoreFile(from string,msg MessageStoreFile) error{

	peer,ok:=s.peers[from]
	if !ok{
		return fmt.Errorf("peer (%s) could not be found in peer list",from)
	}
	
	n,err := s.store.Write(msg.Id,msg.Key,io.LimitReader(peer,msg.Size))
	if err!=nil{
		return err
	}

	fmt.Printf("[%s] written %d bytes to disk: ",s.Transport.Addr(),n)
	peer.CloseStream()
	return nil

}


// bootstrapNetwork is a function that connects to the bootstrap nodes 
// specified in the FileServer options to join the network and discover other peers.
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


// Start starts the file server by listening for incoming connections and handling them
// as well as bootstrapping the network and launching the main loop to process incoming messages and quit signals.
func (s *FileServer) Start() error{

	if err:=s.Transport.ListenAndAccept();err!=nil{
		return err
	}
	s.bootstrapNetwork()
	s.loop()
	return nil
}