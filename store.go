package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	// "io/fs"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "ggnetwork"

func CASPathTransformFunc(key string) PathKey{

	hash:=sha1.Sum([]byte(key))
	hashStr:= hex.EncodeToString(hash[:])

	blocksize:=5
	sliceLen:=len(hashStr)/blocksize
	paths:=make([]string,sliceLen)

	for i:=0;i<sliceLen;i++{

		from,to := i*blocksize,(i*blocksize)+blocksize
		paths[i]=hashStr[from:to]

	}

	return PathKey{

		Pathname: strings.Join(paths,"/"),
		Filename: hashStr,

	}


	// return strings.Join(paths,"/")

}



type PathTransformFunc func(string) PathKey


type PathKey struct{

	Pathname string
	Filename string

}

func (p PathKey) FirstPathName() string{

	paths := strings.Split(p.Pathname,"/")
	if(len(paths)==0){
		return ""
	}

	return paths[0]

}


func (p PathKey) FullPath() string{

	return fmt.Sprintf("%s/%s",p.Pathname,p.Filename)

}

type StoreOpts struct {
	// Root is the folder name of the root,containing all the files/directories
	// of the system
	root string
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(key string) PathKey{
	return PathKey{

		Pathname: key,
		Filename: key,

	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {

	if opts.PathTransformFunc==nil{
		opts.PathTransformFunc=DefaultPathTransformFunc
	}

	if len(opts.root)==0{
		opts.root=defaultRootFolderName
	}

	return &Store{

		StoreOpts: opts,	}

}

func (s *Store) Has(key string) bool{

	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s",s.root,pathKey.FullPath())
	_,err:=os.Stat(fullPathWithRoot)

	if errors.Is(err,os.ErrNotExist){

		return false

	}
	return true
} 


func (s *Store) Clear() error{

	return os.RemoveAll(s.root)

}


func (s *Store) Delete(key string) error{

	pathKey:=s.PathTransformFunc(key)
	FullPathwithRoot := fmt.Sprintf("%s/%s",s.root,pathKey.FirstPathName())

	defer func(){
		log.Printf("Deleted [%s] from disk", pathKey.Filename)
	}()

	return os.RemoveAll(FullPathwithRoot)

}


func (s *Store) Write(key string,r io.Reader) error{

	return s.writeStream(key,r)

}


func (s *Store) Read(key string) (io.Reader, error){

	f,err := s.readStream(key)
	if err!=nil{
		return nil,err
	}

	defer f.Close()

	buf := new(bytes.Buffer)
	_,err = io.Copy(buf,f)

	return buf,err

}


func (s *Store) readStream(key string) (io.ReadCloser, error){

	pathKey := s.PathTransformFunc(key)
	fullPathnameWithRoot := fmt.Sprintf("%s/%s",s.root,pathKey.FullPath())
    return os.Open(fullPathnameWithRoot)

}

func (s *Store) writeStream(key string, r io.Reader) error{
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s",s.root,pathKey.Pathname)
	
	if err:=os.MkdirAll(pathNameWithRoot,os.ModePerm);err!=nil{
		return err
	}

	fullPathNamewithRoot:=fmt.Sprintf("%s/%s",s.root,pathKey.FullPath())	

	f,err := os.Create(fullPathNamewithRoot)
	
	defer f.Close()

	if err!=nil{
		return err
	}

	n,err:=io.Copy(f,r)
	if err!=nil{
		return err
	}

	log.Printf("Written (%d) bytes to disk: %s",n,fullPathNamewithRoot)

	return nil

}