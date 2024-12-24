package main

import (
	"backupusb/archive"
	configuration "backupusb/config"
	"backupusb/crypto"
	"crypto/hmac"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func removePanic(f *os.File, err error) {
	f.Close()
	os.Remove(f.Name())
	panic(err)
}

func main() {
	os.Chdir(filepath.Dir(os.Args[0])) // When started by dragging a file, the working dir would be that of the file
	config := configuration.Load()
	key, err := base64.RawStdEncoding.DecodeString(config.Key)
	if err != nil {
		fmt.Printf("Invalid key in config file: %v", err)
		os.Exit(1)
	}

	startingTime := time.Now()
	switch config.Mode {

	case "encrypt":

		// Parse key
		pubKey, err := x509.ParsePKCS1PublicKey(key)
		if err != nil {
			fmt.Printf("Invalid key in config file. Is it the right one?")
			os.Exit(1)
		}

		// Remove older files
		os.Mkdir("data", os.ModeDir)
		files, err := os.ReadDir("data")
		if err != nil {
			panic(err)
		}
		if len(files) >= config.Backups {
			sort.SliceStable(files, func(i, j int) bool { // Sort reversed
				return files[i].Name() > files[j].Name()
			})

			for _, file := range files[config.Backups-1:] {
				os.Remove("data/" + file.Name())
			}
		}

		// Create file
		encryptedFile, err := os.Create("data/" + strconv.FormatInt(time.Now().UnixMilli(), 10))
		if err != nil {
			panic(err)
		}

		// Create AES writer
		encryptedFile.Write(make([]byte, crypto.ENCRYPTED_HEADER_SIZE)) // Blank header, make space for the future header
		header, mac := crypto.GenHeader()
		hashedWriter := io.MultiWriter(encryptedFile, mac)
		parser, err := crypto.NewAesWriter(header.AesKey, header.IV, hashedWriter)
		if err != nil {
			removePanic(encryptedFile, err)
		}

		// Compress, encrypt and write
		fmt.Println("Compressing...")
		if err = archive.Tar(config.Paths, parser); err != nil {
			removePanic(encryptedFile, err)
		}

		// Flush the remaining buffer and write header to file
		parser.Flush()
		encryptedFile.Seek(0, 0) // Back to start of file
		header.EncryptedMacSum = mac.Sum(nil)
		err = crypto.WriteHeader(encryptedFile, pubKey, &header)
		if err != nil {
			removePanic(encryptedFile, err)
		}
		encryptedFile.Close()

	case "decrypt":

		// Get the target file path
		if len(os.Args) == 1 {
			fmt.Println("Please specify a file path to decrypt")
			os.Exit(1)
		}
		path := strings.Join(os.Args[1:], " ")
		for _, p := range []string{"\"", "'"} {
			path = strings.TrimPrefix(path, p)
			path = strings.TrimSuffix(path, p)
		}

		// Parse key
		privKey, err := x509.ParsePKCS1PrivateKey(key)
		if err != nil {
			fmt.Printf("Invalid key in config file. Is it the right one?")
			os.Exit(1)
		}

		// Open file
		encryptedFile, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer encryptedFile.Close()

		// Read the file header
		header, mac, err := crypto.ReadHeader(encryptedFile, privKey)
		if err != nil {
			panic(err)
		}

		// Verify file integrity
		fmt.Println("Verifying file integrity...")
		if _, err = io.Copy(mac, encryptedFile); err != nil {
			panic(err)
		}
		if !hmac.Equal(header.EncryptedMacSum, mac.Sum(nil)) {
			fmt.Println("Invalid macsum. It seems like the file has been tampered with")
			os.Exit(1)
		}
		encryptedFile.Seek(int64(crypto.ENCRYPTED_HEADER_SIZE), 0) // Back to the encrypted data start

		// Create AES reader
		parser, err := crypto.NewAesReader(header.AesKey, header.IV, encryptedFile)
		if err != nil {
			panic(err)
		}

		// Decrypt and extract
		fmt.Println("Extracting...")
		folderName := filepath.Base(path) + "_"
		os.Mkdir(folderName, os.ModeDir)
		err = archive.Untar(parser, folderName)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Done.")
	fmt.Printf("Execution complete in %v\n", time.Since(startingTime))
}
