package crypto

import (
	"encoding/base64"

	kyberk2so "github.com/symbolicsoft/kyber-k2so"
)

const PUB_KEY_SIZE = kyberk2so.Kyber1024PKBytes
const PRIV_KEY_SIZE = kyberk2so.Kyber1024SKBytes
const SECRET_SIZE = kyberk2so.KyberSSBytes
const CIPHER_SIZE = PUB_KEY_SIZE

func GenKeyPair() (privKey [PRIV_KEY_SIZE]byte, pubKey [PUB_KEY_SIZE]byte) {
	privKey, pubKey, err := kyberk2so.KemKeypair1024()
	if err != nil {
		panic(err)
	}
	return privKey, pubKey
}

func GenParsedKeyPair() (privKey string, pubKey string) {
	sK, pK := GenKeyPair()
	return base64.RawStdEncoding.EncodeToString(sK[:]), base64.RawStdEncoding.EncodeToString(pK[:])
}

func GetSharedKey(pubKey [PUB_KEY_SIZE]byte) (chiper [CIPHER_SIZE]byte, secret [SECRET_SIZE]byte) {
	chiper, secret, err := kyberk2so.KemEncrypt1024(pubKey)
	if err != nil {
		panic(err)
	}
	return chiper, secret
}

func DecryptKey(cipher [CIPHER_SIZE]byte, privKey [PRIV_KEY_SIZE]byte) (secret [SECRET_SIZE]byte) {
	secret, err := kyberk2so.KemDecrypt1024(cipher, privKey)
	if err != nil {
		panic(err)
	}
	return secret
}

func ParseDecrypt(cipher, privKey []byte) []byte {
	r := DecryptKey([CIPHER_SIZE]byte(cipher), [PRIV_KEY_SIZE]byte(privKey))
	return r[:]
}
