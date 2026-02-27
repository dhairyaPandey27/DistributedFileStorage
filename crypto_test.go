package main

import (
	"bytes"
	// "fmt"
	"testing"
)

func TestCopyEncrypt(t *testing.T) {

	payload := "FOO NOT BAR"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := newEncryptionKey()
	_,err:=copyEncrypt(key,src,dst)
	if err!=nil{
		t.Error(err)
	}
	
	out := new(bytes.Buffer)
	if _,err := copyDecrypt(key,dst,out);err!=nil{
		t.Error(err)
	}

	if(out.String()!=payload){
		t.Errorf("Decryption failed")
	}

}