package main


import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)


const defaultRootFolderName = "ggnetwork"


// PathKey is a struct the represents the pathname and filename of a file.
type PathKey struct{

	Pathname string
	Filename string

}


// PathTransformFunc is a type of function that takes a 
// string key and return a Pathkey struct.
type PathTransformFunc func(string) PathKey


// CASPathTransformFunc is a default implementation of the PathTransformFunc
// that transforms a given key into a PathKey. It uses hashing to generate a
// hash string which is fragmented in fixed block size to get a CAS path.
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

}


// FirstPathName returns the first segment of the Pathname
//  which is the first directory in the path.
func (p PathKey) FirstPathName() string{

	paths := strings.Split(p.Pathname,"/")
	if(len(paths)==0){
		return ""
	}

	return paths[0]

}


// FullPath returns the complete path by combining the Pathname and Filename.
func (p PathKey) FullPath() string{

	return fmt.Sprintf("%s/%s",p.Pathname,p.Filename)

}


// DefaultPathTransformFunc is type of PathTransformFunc that
// return the default values for Pathname and Filename
var DefaultPathTransformFunc = func(key string) PathKey{
	return PathKey{
		
		Pathname: key,
		Filename: key,
		
	}
}


// StoreOpts is a struct that contains configuration options for the Store.
type StoreOpts struct {
	// root is the folder name of the root,containing all the files/directories
	// of the system
	root string
	// ID of the owner of the storage, which will we be used to store all files at that location
	// so we can sync all the files if needed
	ID string
	PathTransformFunc PathTransformFunc
}


// Store is a struct that is used for storing and retrieving files
// and it used the StoreOpts for configuration.
type Store struct {
	StoreOpts
}


// NewStore creates a new instance of Store with the provided options
// and also sets default values for the root folder name, ID
// and PathTransformFunc if they are not provided in the options.
func NewStore(opts StoreOpts) *Store {

	if opts.PathTransformFunc==nil{
		opts.PathTransformFunc=DefaultPathTransformFunc
	}

	if len(opts.root)==0{
		opts.root=defaultRootFolderName
	}

	if len(opts.ID)==0{
		opts.ID=generateID()
	}

	return &Store{

		StoreOpts: opts,	}

}


// Has checks if a file with the given key exists for the specified peer ID in the storage.
func (s *Store) Has(id string,key string) bool{

	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s",s.root,id,pathKey.FullPath())
	_,err:=os.Stat(fullPathWithRoot)

	if errors.Is(err,os.ErrNotExist){

		return false

	}
	return true
} 


// Clear removes all files and directories in the root folder of the storage, effectively clearing the storage.
func (s *Store) Clear() error{

	return os.RemoveAll(s.root)

}


// Delete removes the file associated with the given key for the specified peer ID from the storage.
func (s *Store) Delete(id string,key string) error{

	pathKey:=s.PathTransformFunc(key)
	FullPathwithRoot := fmt.Sprintf("%s/%s/%s",s.root,id,pathKey.FirstPathName())

	defer func(){
		log.Printf("Deleted [%s] from disk", pathKey.Filename)
	}()

	return os.RemoveAll(FullPathwithRoot)

}


func (s *Store) Write(id string,key string,r io.Reader) (int64,error){

	return s.writeStream(id,key,r)

}


func (s *Store) Read(id string,key string) (int64,io.Reader, error){

	return s.readStream(id,key)

}


// readStream reads the file associated with the given key for the
// specified peer ID from the storage and returns its size and a reader for the file.
func (s *Store) readStream(id string,key string) (int64,io.ReadCloser, error){

	pathKey := s.PathTransformFunc(key)
	fullPathnameWithRoot := fmt.Sprintf("%s/%s/%s",s.root,id,pathKey.FullPath())
   	file,err := os.Open(fullPathnameWithRoot)
	if err!=nil{
		return 0,nil,err
	}

	fi,err:= file.Stat()
	if err!=nil{
		return 0,nil,err
	}

	return fi.Size(),file,nil

}


// writeDecrypt reads encrypted data from the provided reader, decrypts it using the provided encryption key
func (s *Store) writeDecrypt(id string,encKey []byte,key string, r io.Reader) (int64,error){

	f,err:= s.openFileForWriting(id,key)
	if err!=nil{
		return 0,err
	}

	n,err:= copyDecrypt(encKey,r,f)
	return int64(n),err


}


// openFileForWriting is a helper function that creates and opens a file 
// for writing based on the given peer ID and key, and returns the file handle.
func (s *Store) openFileForWriting(id string,key string) (*os.File,error){

	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s",s.root,id,pathKey.Pathname)

	if err:=os.MkdirAll(pathNameWithRoot,os.ModePerm);err!=nil{
		return nil,err
	}

	fullPathNamewithRoot:=fmt.Sprintf("%s/%s/%s",s.root,id,pathKey.FullPath())	

	return os.Create(fullPathNamewithRoot)

}


// writeStream writes data from the provided reader to a file associated with the given key for the specified peer ID in the storage.
func (s *Store) writeStream(id string,key string, r io.Reader) (int64,error){

	f,err:=s.openFileForWriting(id,key)
	if err!=nil{
		return 0,err
	}

	return io.Copy(f,r)

}