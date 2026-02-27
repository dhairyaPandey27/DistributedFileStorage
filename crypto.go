package main


import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)


// generateID creates a random 32-byte ID and returns it as a hexadecimal string.
func generateID() string{

	buf:=make([]byte,32)
	io.ReadFull(rand.Reader,buf)
	return hex.EncodeToString(buf)

}


// hashKey takes a key and using the md5 hashing technique,
// it returns the hash of the key as a hexadecimal string.
func hashKey(key string) string{
	hash:=md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}


// newEncryptionKey generates a new random 32-byte encryption key.
func newEncryptionKey() []byte{

	keyBuf:=make([]byte,32)
	io.ReadFull(rand.Reader,keyBuf)
	return keyBuf

}


// copyStream reads data from the source and processes it using the 
// provided cipher stream, and writes the processed data to the destination.
func copyStream(stream cipher.Stream, blockSize int, src io.Reader,dst io.Writer) (int, error){

	var (
			buf    = make([]byte, 32*1024)
			nw = blockSize
		)

		for {
			n, err := src.Read(buf) // Read data into a buffer
			if n > 0 {
				// This will generate a random aes encrypted data called keystream,
				// then will xor the data present in buf with the keystream and
				// store the result back into buf
				stream.XORKeyStream(buf, buf[:n]) 

				// Store the xor value in dest
				nn, err := dst.Write(buf[:n])
				if err != nil {
					return 0, err
				}
				nw+=nn
			}

			if err == io.EOF {
				break
			}

			if err != nil {
				return 0, err
			}

		}
	return nw,nil
}


// copyDecrypt reads encrypted data from the source, decrypts it using the provided key 
// and writes the decrypted data to the destination using copyStream function.
func copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error){

	block, err :=aes.NewCipher(key)
	if err!=nil{
		return 0,err
	}

	iv:=make([]byte, block.BlockSize())

	if _,err:= src.Read(iv);err!=nil{
		return 0,err
	}

	stream := cipher.NewCTR(block,iv)
	return copyStream(stream,block.BlockSize(),src,dst)

}


// copyEncrypt reads data from the source, encrypts it using the 
// provided key and writes the encrypted data to the destination using copyStream function.
func copyEncrypt(key []byte, src io.Reader, dest io.Writer) (int, error) {
	// Create a new block using your key
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	// Create a initialisation vector, which is of same size of block
	iv := make([]byte, block.BlockSize())

	// Fill iv with a random value
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	// Write iv to destination so that it stays at the beginning of the
	// destination and can be used by decryption 
	if _, err := dest.Write(iv); err != nil {
		return 0, err
	}
	
	stream := cipher.NewCTR(block,iv)
	return copyStream(stream,block.BlockSize(),src,dest)

}