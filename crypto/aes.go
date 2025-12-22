package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io"
)

func getStream(aesKey, iv []byte) (cipher.Stream, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() {
		return nil, errors.New("invalid IV length")
	}

	return cipher.NewCTR(block, iv), nil
}

func NewAesWriter(aesKey, iv []byte, out io.Writer) (io.WriteCloser, error) {
	stream, err := getStream(aesKey, iv);
	if err != nil {
		return nil, err
	}

	return &cipher.StreamWriter{S: stream, W: out}, nil
}

func NewAesReader(aesKey, iv []byte, in io.Reader) (io.Reader, error) {
	stream, err := getStream(aesKey, iv);
	if err != nil {
		return nil, err
	}

	return &cipher.StreamReader{S: stream, R: in}, nil
}
