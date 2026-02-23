package main

import (
	"fmt"
	"log"

	"github.com/dhairyaPandey27/DistributedFileStorage/p2p"
)

type FileServerOpts struct {

	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes  []string
}

type FileServer struct{	
	FileServerOpts

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
	}

}

func (s *FileServer) Stop(){

	close(s.quitch)

}

func (s *FileServer) loop(){

	defer func(){

		log.Println("file server stopped due to user action")
		s.Transport.Close()

	}()

	for{

		select{

		case msg:= <-s.Transport.Consume():	
			fmt.Println(msg)
		case <- s.quitch:
			return

		}

	}

}

func (p *FileServer) bootstrapNetwork() error{

	for _,addr := range p.BootstrapNodes{

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