package backups

import (
	"backupusb/archive"
	"backupusb/crypto"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
)

func DecryptBackup(path, destination string, privKey []byte, extract bool) (uint64, uint64) {

	// Open file
	inFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	// Read the macsum
	macSum := make([]byte, crypto.MACSUM_SIZE)
	if _, err := io.ReadFull(inFile, macSum); err != nil {
		panic(err)
	}

	// Read the file header (keys)
	header, mac, err := crypto.ReadHeader(inFile, privKey)
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
	aesReader, err := crypto.NewAesReader(header.AesKey, header.IV, inFile)
	if err != nil {
		panic(err)
	}

	// Set up decompression (don't extract yet)
	zstdReader, err := zstd.NewReader(aesReader, zstd.WithDecoderConcurrency(1)) // No need to specify compression level. If concurrency is enabled (>1) the end of the file isn't copied
	if err != nil {
		panic(err)
	}
	defer zstdReader.Close()

	// Decrypt only
	if !extract {
		outFile, err := os.Create(filepath.Join(filepath.Base(path) + ".tar"))
		if err != nil {
			panic(err)
		}
		defer outFile.Close()

		io.Copy(outFile, zstdReader)
		return 1, 0
	}

	// Decrypt and extract
	fmt.Println("Extracting...")
	folderName := filepath.Join(destination, "_"+filepath.Base(path))
	os.Mkdir(folderName, os.ModePerm)

	fileN, folderN, err := archive.Untar(zstdReader, folderName)
	header.Destroy()
	if err != nil {
		panic(err)
	}
	return fileN, folderN
}
