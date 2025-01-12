package main

import (
	"backupusb/backups"
	"backupusb/configuration"
	"backupusb/crypto"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var usageMsgs = map[string]string{
	"help":    "help",
	"config":  "config",
	"decrypt": "decrypt <file> [destination] [--tar]",
}

const invalidConfigMsg = "Invalid config file. Please delete it and generate a new one"

func parsePath(path string) string {
	for _, p := range []string{"\"", "'"} {
		path = strings.TrimPrefix(path, p)
		path = strings.TrimSuffix(path, p)
	}
	return strings.ReplaceAll(path, "\\", "/")
}

func showHelp() {
	s := strings.Repeat(" ", 4)

	fmt.Printf("Usage: %s [help | config | decrypt]\n\n", os.Args[0])
	fmt.Printf("  * %s %s\n%s - Showes you this message\n\n", os.Args[0], usageMsgs["help"], s)
	fmt.Printf("  * %s %s\n%s - Lets you edit the program configuration\n\n", os.Args[0], usageMsgs["config"], s)

	fmt.Printf(
		"  * %s %s\n%s - Decrypts a previous backup file\n"+
			"%s - You can also set the private key as an enviroment variable (PRIV_KEY) to avoid pausing\n"+
			"%s - Please AVOID storing the key as a persistent value and only set it on each execution\n",
		os.Args[0], usageMsgs["decrypt"], s, s, s,
	)
}

func main() {
	if !run() {
		showHelp()
	}
}

func run() bool {
	args := os.Args[1:]

	// * Start backup
	if len(args) == 0 {
		config, err := configuration.Load()
		if err != nil {
			fmt.Println(invalidConfigMsg)
			os.Exit(1)
		}

		// Decode pubKey
		pubKey, err := base64.RawStdEncoding.DecodeString(config.Key)
		if err != nil {
			fmt.Println("Invalid key in config file")
			os.Exit(1)
		}

		// Verify key len
		if len(pubKey) != crypto.PUB_KEY_SIZE {
			fmt.Println("Invalid key in config file. Is it the right one?")
			os.Exit(1)
		}

		// Remove older files
		startingTime := time.Now()
		os.Mkdir(config.Destination, os.ModeDir)
		backups.DeleteOldBackups(config.Destination, config.Amount)

		// Create file
		outFile, err := os.Create(config.Destination + strconv.FormatInt(startingTime.UnixMilli(), 10))
		if err != nil {
			panic(err)
		}
		defer outFile.Close()

		// Backup to file
		fileN, folderN := backups.CreateBackup(outFile, [crypto.PUB_KEY_SIZE]byte(pubKey), config.Paths)
		crypto.DestroyKey(pubKey)

		fmt.Println("\nDone.")
		fmt.Printf("%d files and %d folders have been affected\n", fileN, folderN)
		fmt.Printf("Execution completed in %v\n", time.Since(startingTime).Round(time.Millisecond))
		return true
	}

	switch args[0] {

	case "config":
		if len(args) != 1 {
			fmt.Println("Usage:", os.Args[0], usageMsgs["config"])
			os.Exit(1)
		}

		// Load/Create the config file
		config, err := configuration.Load()
		if err != nil {
			fmt.Println(invalidConfigMsg)
			os.Exit(1)
		}

		// This wont run if the config has just been created
		config.OpenEditor()
		return true

	case "decrypt":
		usageMsg := "Usage: " + os.Args[0] + " " + usageMsgs["decrypt"]
		args = args[1:]
		if len(args) == 0 {
			fmt.Println(usageMsg)
			os.Exit(1)
		}

		// Should it extract the file to a folder
		extract := true
		if args[len(args)-1] == "--tar" {
			extract = false
			args = args[:len(args)-1]
		}
		if args[0] == "--tar" {
			extract = false
			args = args[1:]
		}

		// Validate the arguments
		if len(args) == 0 || len(args) > 2 {
			fmt.Println(usageMsg)
			return true
		}
		os.Chdir(filepath.Dir(os.Args[0]))
		target := parsePath(args[0])

		// Get the destination
		destination := filepath.Dir(os.Args[0])
		if len(args) == 2 {
			destination = parsePath(args[1])
		}

		// Ask for the private key
		b64PrivKey := strings.Trim(os.Getenv("PRIV_KEY"), " ") // Checks if stored
		if b64PrivKey == "" {
			fmt.Print("Enter the private/decryption key: ")
			fmt.Scanln(&b64PrivKey)
		}

		// Verifies it
		privKey, err := base64.RawStdEncoding.DecodeString(b64PrivKey)
		if err != nil || len(privKey) != crypto.PRIV_KEY_SIZE {
			fmt.Println("Invalid key")
			os.Exit(1)
		}
		crypto.DestroyKeyString(&b64PrivKey) // Works poorly but it's not really required, so we'll leave it here

		// Decrypt the backup
		startingTime := time.Now()
		fileN, folderN := backups.DecryptBackup(target, destination, privKey, extract)
		crypto.DestroyKey(privKey)

		fmt.Println("Done.")
		if extract {
			fmt.Printf("%d files and %d folders have been affected\n", fileN, folderN)
		}
		fmt.Printf("Execution completed in %v\n", time.Since(startingTime).Round(time.Millisecond))
		return true
	}

	// No need for an "help" command, since it runs by default

	return false
}
