package config

import (
	"backupusb/crypto"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const CONFIG_PATH = "config.json"

type Config struct {
	Mode    string   `json:"mode"`
	Key     string   `json:"key"`
	Paths   []string `json:"paths"`
	Backups int      `json:"backups"`
}

func Create() {
	f, err := os.Create(CONFIG_PATH)
	if err != nil {
		panic(err)
	}

	// Generate default config
	config, err := json.Marshal(Config{
		Mode:    "encrypt",
		Backups: 5,
		Paths:   []string{},
	})
	if err != nil {
		panic(err)
	}
	f.Write(config)
	f.Close()
}

func Load() *Config {
	confFile, err := os.Open(CONFIG_PATH)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}

		Create()
		fmt.Println("Please fill the config.json file and run again")
		privKey, pubKey := crypto.GenParsedKeyPair()
		fmt.Println("Here are some sample keys:\n\n - Private key:", privKey, "\n\n - Public key:", pubKey, "\n\n Press ENTER to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(0)
	}

	data, err := io.ReadAll(confFile)
	if err != nil {
		panic(err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Invalid config: %v\n", err)
		os.Exit(1)
	}

	return &config
}
