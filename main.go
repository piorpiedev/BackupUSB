package main

import (
	"backupusb/archive"
	configuration "backupusb/config"
	"backupusb/crypto"
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

const FILE_EXTENSION = ".bk"

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

	var fileN, folderN uint64 = 0, 0
	startingTime := time.Now()
	switch config.Mode {

	case "encrypt":
		// Verify key len
		if len(key) != crypto.PUB_KEY_SIZE {
			fmt.Printf("Invalid key in config file. Is it the right one?")
			os.Exit(1)
		}

		// Remove older files
		os.Mkdir("data", os.ModeDir)
		files, err := os.ReadDir("data")
		if err != nil {
			panic(err)
		}
		if config.Backups < -1 || config.Backups == 0 {
			fmt.Printf("Invalid backups value in config file: %v", config.Backups)
			fmt.Printf("Set it to a positive integer to limit the amount of backups in the data folder, or you can set it -1 to disable this feature")
			os.Exit(1)
		}
		if config.Backups != -1 && len(files) >= config.Backups {
			sort.SliceStable(files, func(i, j int) bool { // Sort reversed
				return files[i].Name() > files[j].Name()
			})

			for _, file := range files[config.Backups-1:] {
				os.Remove("data/" + file.Name())
			}
		}

		// Create file
		outFile, err := os.Create("data/" + strconv.FormatInt(startingTime.UnixMilli(), 10) + FILE_EXTENSION)
		if err != nil {
			panic(err)
		}
		defer outFile.Close()

		// Create AES writer
		outFile.Write(make([]byte, crypto.MACSUM_SIZE)) // Make space for the future macsum
		header, enHeader := crypto.GenHeader([crypto.PUB_KEY_SIZE]byte(key))
		mac := crypto.NewMAC(header.MacKey)
		parser := io.MultiWriter(outFile, mac)
		enWriter, err := crypto.NewAesWriter(header.AesKey, header.IV, parser)
		if err != nil {
			removePanic(outFile, err)
		}

		// Write header
		_, err = parser.Write(enHeader.Dump())
		if err != nil {
			removePanic(outFile, err)
		}

		// Compress, encrypt and write
		// While the compression is executed, the header memory is generally reused (since it's not called anymore by the code)
		fmt.Println("Compressing...")
		if fileN, folderN, err = archive.Tar(config.Paths, enWriter); err != nil {
			removePanic(outFile, err)
		}

		// Flush the remaining buffer and write the macsum at the start of the file
		enWriter.Flush()
		outFile.Seek(0, 0)
		msum := mac.Sum(nil)
		outFile.Write(msum) // Write the 64 bytes of encrypted macsum (this writes to the mac too, but we already evaluated the sum)

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

		// Verify key len
		if len(key) != crypto.PRIV_KEY_SIZE {
			fmt.Println("Invalid key in config file. Is it the right one?")
			os.Exit(1)
		}

		// Open file
		inFile, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer inFile.Close()

		// Read the macsum
		macSum := make([]byte, crypto.MACSUM_SIZE)
		inFile.Read(macSum)

		// Read the file header (keys)
		header, mac, err := crypto.ReadHeader(inFile, key)
		if err != nil {
			panic(err)
		}

		// Verify file integrity
		verStartTime := time.Now()
		fmt.Println("Verifying file integrity...")
		if _, err = io.Copy(mac, inFile); err != nil {
			panic(err)
		}
		if !crypto.CompareMacSums(macSum, mac.Sum(nil)) {
			fmt.Printf("Invalid macsum. It seems like the file has been tampered with (%v)\n", time.Since(verStartTime))
			os.Exit(1)
		}
		fmt.Printf("Integrity verified in %v\n\n", time.Since(verStartTime))
		inFile.Seek(int64(crypto.ENCRYPTED_HEADER_SIZE+crypto.MACSUM_SIZE), 0) // Back to the encrypted data start

		// Create AES reader
		enReader, err := crypto.NewAesReader(header.AesKey, header.IV, inFile)
		if err != nil {
			panic(err)
		}

		// Decrypt and extract
		fmt.Println("Extracting...")
		folderName := strings.TrimSuffix(filepath.Base(path), FILE_EXTENSION)
		os.Mkdir(folderName, os.ModeDir)
		fileN, folderN, err = archive.Untar(enReader, folderName)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("\nDone.")
	fmt.Printf("%v files and %v folders have been affected\n", fileN, folderN)
	fmt.Printf("Execution completed in %v\n", time.Since(startingTime))
}
