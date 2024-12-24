package crypto

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha512"
	"errors"
	"hash"
	"io"
)

const MACSUM_SIZE = 64 // AES, so EncryptedSize = Size
const IV_SIZE = 16     // The secret is of size 32, but we only want the first 16
const ENCRYPTED_HEADER_SIZE = CIPHER_SIZE*3 + MACSUM_SIZE

type EncryptedHeader struct {
	AesKey          []byte
	SharedIV        []byte
	MacKey          []byte
	EncryptedMacSum []byte // This is encrypted in AES
}
type Keys struct {
	AesKey []byte
	IV     []byte
	MacKey []byte
	MacSum [MACSUM_SIZE]byte
}

// * Decrypt

func ParseHeader(in []byte) (*Keys, []byte, hash.Hash, error) {
	if len(in) != ENCRYPTED_HEADER_SIZE {
		return nil, []byte{}, nil, errors.New("invalid header")
	}

	sharedIV := in[SECRET_SIZE : SECRET_SIZE*2]
	enMacSum := in[SECRET_SIZE*3:]

	header := Keys{
		AesKey: in[:SECRET_SIZE],
		IV:     sharedIV[:16],
		MacKey: in[SECRET_SIZE*2 : SECRET_SIZE*3],
	}

	mac := hmac.New(sha512.New, header.MacKey)
	mac.Write(header.AesKey)
	mac.Write(sharedIV)

	return &header, enMacSum, mac, nil
}

// * Encrypt

func GenKeys() (*Keys, *EncryptedHeader, hash.Hash, [PRIV_KEY_SIZE]byte) {
	privKey, pubKey := GenKeyPair()

	// en_ indica i valori cryptati
	enAesKey, aesKey := GetSharedKey(pubKey)
	enMacKey, macKey := GetSharedKey(pubKey)
	enIV, shIV := GetSharedKey(pubKey) // We have a shared value, as kyber only allowes for 32 bytes secrets, but we use 16 bytes IVs

	mac := hmac.New(sha512.New, macKey[:])
	mac.Write(aesKey[:])
	mac.Write(shIV[:])

	keys := &Keys{
		AesKey: aesKey[:],
		MacKey: macKey[:],
		IV:     shIV[:16],
	}

	enHeader := &EncryptedHeader{
		AesKey:   enAesKey[:],
		MacKey:   enMacKey[:],
		SharedIV: enIV[:],
	}

	return keys, enHeader, mac, privKey
}

func (h *EncryptedHeader) Dump() []byte {
	b := make([]byte, 0, ENCRYPTED_HEADER_SIZE)
	b = append(b, h.AesKey...)
	b = append(b, h.SharedIV...)
	b = append(b, h.MacKey...)
	b = append(b, h.EncryptedMacSum...)

	return b
}

// ------- //

func WriteHeader(out io.Writer, publicKey []byte, header *Header) error {
	hd := header.Dump()
	encryptedHeader, err := RSAEncrypt(hd, key)
	if err != nil {
		return err
	}
	out.Write(encryptedHeader) // Write header

	return nil
}

func ReadHeader(in io.Reader, key *rsa.PrivateKey) (*Header, hash.Hash, error) {
	// Read encrypted header
	encryptedHeader := make([]byte, ENCRYPTED_HEADER_SIZE)
	_, err := in.Read(encryptedHeader)
	if err != nil {
		return nil, nil, err
	}

	// Decrypt the header
	h, err := RSADecrypt(encryptedHeader, key)
	if err != nil { // Invalid key or content
		return nil, nil, err
	}

	return ParseHeader(h)
}
