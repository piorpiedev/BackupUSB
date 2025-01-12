package configuration

import (
	"backupusb/crypto"
	"bufio"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"os"
)

const CONFIG_PATH = "config.bc"
const DEFAULT_DESTINATION = "data/"
const MIN_AMOUNT = 5

func getConfigKey() (key, iv []byte) {
	// These two are just static values used for static encryption.
	// They are meant to be left as static values inside the binary and will only add a SMALL layer of security through obscurity.
	key, _ = base64.RawStdEncoding.DecodeString("u4ytUHYlZ48Vmd7AMBvPvJ7QNhUPsM0C7ZYfdSxwdLM")
	iv, _ = base64.RawStdEncoding.DecodeString("gwLEoAj6MvcjuihlvdzvNg")
	return key, iv
}

type Config struct {
	Key         string   // The public key
	Paths       []string // A list of folders/files to backup
	Amount      int      // Max amount of backups to store (oldest deleted first, set to -1 to disable)
	Destination string   // The folder where the backups are stored
}

func (c *Config) Save() error {
	confFile, err := os.Create(CONFIG_PATH)
	if err != nil {
		return err
	}
	defer confFile.Close()

	// Pass through encryptor
	key, iv := getConfigKey()
	aesWriter, err := crypto.NewAesWriter(key, iv, confFile)
	defer aesWriter.Flush()
	if err != nil {
		panic(err)
	}

	// Serialize data
	enc := gob.NewEncoder(aesWriter)
	if err := enc.Encode(c); err != nil {
		panic(err)
	}

	crypto.DestroyKey(key)
	crypto.DestroyKey(iv)
	return nil
}

func createDefault(pubKey string) *Config {

	// Generate default config
	config := Config{
		Key:         pubKey,
		Amount:      5,
		Paths:       []string{},
		Destination: DEFAULT_DESTINATION,
	}

	config.Save()
	return &config
}

func Load() (*Config, error) {
	confFile, err := os.Open(CONFIG_PATH)

	// No config file found
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}

		// Generate default config
		privKey, pubKey := crypto.GenParsedKeyPair()
		config := createDefault(pubKey)
		fmt.Printf(
			"No config file found, so a default one has been created\nThis is a brand new randly generated private/decryption key: \n\n%s\n\n"+
				"The public key has already been added to the config file. Please take note of the private key, and store it in a safe place\n"+
				"To edit the config in the future, you can simply run: %s config\n\n"+
				" - Press ENTER to continue editing the configuration...", privKey, os.Args[0],
		)
		bufio.NewReader(os.Stdin).ReadBytes('\n') // Pause console

		config.OpenEditor()
		fmt.Println(
			"All done. Next time you run the program, it will start backing up with the new configuration., ",
			"\nYou can also edit the config again by using:", os.Args[0], "config",
		)
		os.Exit(0)
	}
	defer confFile.Close()

	// Pass through decryptor
	key, iv := getConfigKey()
	aesReader, err := crypto.NewAesReader(key, iv, confFile)
	if err != nil {
		panic(err)
	}

	// Deserialize data
	var config Config
	dec := gob.NewDecoder(aesReader)
	if err := dec.Decode(&config); err != nil {
		return nil, err
	}

	crypto.DestroyKey(key)
	crypto.DestroyKey(iv)
	return &config, nil
}
