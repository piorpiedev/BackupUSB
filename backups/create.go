package backups

import (
	"backupusb/archive"
	"backupusb/crypto"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/klauspost/compress/zstd"
)

func DeleteOldBackups(folderPath string, amount int) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		panic(err)
	}

	if amount < -1 || amount == 0 {
		fmt.Printf("Invalid backups value in config file: %d", amount)
		fmt.Printf("Set it to a positive integer to limit the amount of backups in the data folder, or you can set it -1 to disable this feature")
		os.Exit(1)
	}

	if amount != -1 && len(files) >= amount {
		sort.SliceStable(files, func(i, j int) bool { // Sort reversed
			return files[i].Name() > files[j].Name()
		})

		for _, file := range files[amount-1:] {
			os.Remove(folderPath + file.Name())
		}
	}
}

func removePanic(f *os.File, err error) {
	f.Close()
	os.Remove(f.Name())
	panic(err)
}

func CreateBackup(outFile *os.File, pubKey [crypto.PUB_KEY_SIZE]byte, paths []string) (fileN uint64, folderN uint64) {
	outFile.Write(make([]byte, crypto.MACSUM_SIZE)) // Make space for the future macsum
	header, enHeader := crypto.GenHeader(pubKey)

	// Prepare the writers
	mac := crypto.NewMAC(header.MacKey)
	macAndFile := io.MultiWriter(outFile, mac)
	aesWriter, err := crypto.NewAesWriter(header.AesKey, header.IV, macAndFile)
	if err != nil {
		removePanic(outFile, err)
	}

	// Write header
	if _, err := macAndFile.Write(enHeader.Dump()); err != nil {
		removePanic(outFile, err)
	}

	// Compress, encrypt and write
	fmt.Println("Compressing...")
	zstdWriter, err := zstd.NewWriter(aesWriter, zstd.WithEncoderLevel(zstd.SpeedBetterCompression), zstd.WithEncoderConcurrency(4)) // Compression level 5. Apprently we can keep concurrency here
	if err != nil {
		removePanic(outFile, err)
	}
	if fileN, folderN, err = archive.Tar(paths, zstdWriter); err != nil {
		removePanic(outFile, err)
	}
	zstdWriter.Close()

	// Flush the remaining buffer and write the macsum at the start of the file
	outFile.Seek(0, 0)
	msum := mac.Sum(nil)
	header.Destroy()
	outFile.Write(msum) // Write the 64 bytes of encrypted macsum (this writes to the mac too, but we already evaluated the sum)

	return fileN, folderN
}
