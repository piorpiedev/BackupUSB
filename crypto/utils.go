package crypto

func DestroyKey(key []byte) {
	for i := range key {
		key[i] = 0
	}
}

func DestroyKeyString(key *string) {
	DestroyKey([]byte(*key))
	*key = ""
}
