package main


import (
	"bytes"
	"io"
	"testing"
)


func TestPathTransformFunc(t *testing.T){

	key:="momsbestpicture"
	pathKey:=CASPathTransformFunc(key)

	expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathKey.Pathname!=expectedPathName{
		t.Errorf("have %s want %s",pathKey.Pathname,expectedPathName)
	}

}


func TestStore(t *testing.T){

	s:=newStore()
	defer teardown(t,s)

	id:=generateID()

	key:="momsspecial"

	data := []byte("some jpeg bytes")
	if _,err := s.Write(id,key,bytes.NewReader(data));err!=nil{
		t.Error(err)
	}

	if ok:=s.Has(id,key);!ok{
		t.Errorf("Expected to have key %s",key)
	}

	_,r,err:=s.Read(id,key)
	if err!=nil{
		t.Error(err)
	}

	b,_:=io.ReadAll(r)
	if(string(b)!=string(data)){
		t.Errorf("want %s have %s",data,b)
	}

	if err:=s.Delete(id,key);err!=nil{
		t.Error(err)
	}

	if ok:=s.Has(id,key);ok{
		t.Errorf("Expected to not have key %s",key)
	}



}


func newStore() *Store{

	opts:=StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	return NewStore(opts)

}


func teardown(t *testing.T, s *Store){

	if err := s.Clear(); err!=nil{
		t.Error(err)
	}

}