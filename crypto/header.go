package crypto

import (
	"errors"
	"hash"
	"io"
)

const IV_SIZE = 16 // MUST BE < CIPHER_SIZE | The secret is 32B, but we only take the first 16B
const ENCRYPTED_HEADER_SIZE = CIPHER_SIZE * 3

type EncryptedHeader Header
type Header struct {
	AesKey []byte
	IV     []byte
	MacKey []byte
}

// * Decrypt

func ParseHeader(in []byte) (*EncryptedHeader, error) {
	if len(in) != ENCRYPTED_HEADER_SIZE {
		return nil, errors.New("invalid header")
	}

	header := EncryptedHeader{
		AesKey: in[:CIPHER_SIZE],
		IV:     in[CIPHER_SIZE : CIPHER_SIZE*2],
		MacKey: in[CIPHER_SIZE*2:],
	}

	return &header, nil
}

func (h *EncryptedHeader) DecryptKeys(privateKey []byte) *Header {
	return &Header{
		AesKey: ParseDecrypt(h.AesKey, privateKey),
		IV:     ParseDecrypt(h.IV, privateKey)[:IV_SIZE],
		MacKey: ParseDecrypt(h.MacKey, privateKey),
	}
}

// * Encrypt

func GenHeader(pubKey [PUB_KEY_SIZE]byte) (*Header, *EncryptedHeader) {
	// en_ indica i valori cryptati
	enAesKey, aesKey := GetSharedKey(pubKey)
	enMacKey, macKey := GetSharedKey(pubKey)
	enIV, shIV := GetSharedKey(pubKey) // We have a shared value, as kyber only allowes for 32 bytes secrets, but we use 16 bytes IVs

	header := &Header{
		AesKey: aesKey[:],
		MacKey: macKey[:],
		IV:     shIV[:IV_SIZE],
	}

	enHeader := &EncryptedHeader{
		AesKey: enAesKey[:],
		MacKey: enMacKey[:],
		IV:     enIV[:],
	}

	return header, enHeader
}

func (h *EncryptedHeader) Dump() []byte {
	b := make([]byte, 0, ENCRYPTED_HEADER_SIZE)
	b = append(b, h.AesKey...)
	b = append(b, h.IV...)
	b = append(b, h.MacKey...)

	return b
}

func ReadHeader(in io.Reader, privKey []byte) (*Header, hash.Hash, error) {
	// Read encrypted header
	data := make([]byte, ENCRYPTED_HEADER_SIZE)
	if _, err := in.Read(data); err != nil {
		return nil, nil, err
	}

	enHeader, err := ParseHeader(data)
	if err != nil { // Invalid header
		return nil, nil, err
	}

	header := enHeader.DecryptKeys(privKey)
	mac := NewMAC(header.MacKey)
	mac.Write(data)

	return header, mac, nil
}
