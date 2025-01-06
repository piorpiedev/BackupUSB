# BackupUSB

This is a simple project, meant to safely backup to an external drive one ore multiple paths, recursively, in order for them to be restored later
This project is MEANT to rely on asyncronous encryption, in order to keep the backups (stored with the program to generate them) and the decryption key separately

---

# Algorithms

  - Sha512: Used for the MacSum, in order to verify the file integrity before decrypting
  - AES256 CTR: Used to encrypt the main data block, using a random key generated on every execution
  - Crystals Kyber K2SO: Used to generated and safeley encrypt the whole header, key by key

---

# How this works

## First run

> The program gets first executed, creates an empty config, and suggests to the user a newly generated key pair

## Encryption

> Backups the exceed the backup amount get deleted (oldest first) | Set it to -1 to disable
>
> The file is created (`data/[TIMESTAMP].bk`), and 64 bytes at the start are skipped for the macsum
>
> The preencrypted header is written to file, as well as the data itself, that gets encrypted as the same time as it's archived (in order to avoid any possible file recovery)
>
> The macsum of the rest of the file (both encrypted header AND data) is finally written at the start of the file

## Decryption

> The MacSum is read, followed by the header
>
> We evalutate the MacSum of the encrypted header and data and compare it to the MacSum found previously
>
> IF, and only if, it matches, continue with the extraction into a folder named with the backup timestamp

---

# File Structure

`[MacSum]` | `[AesKey]` `[IV]` `[MacKey]` | `[Data]`

### First Block (MacSum, Plain/Sha512)

  - **[MacSum]**: 64B - Sha512 of the already encrypted file, in order (both header and data)

### Second Block (Header, Crystal)

  - **[AesKey]***: 1568B / 32B - Random key generated with Crystals Kyber
  - **[IV]***: 1568B / 16B - Random key generated with Crystals Kyber (NOTE: On decryption this returns 32B, we take the first 16 of those)
  - **[MacKey]***: 1568B / 32B - Random key generated with Crystals Kyber

### Third Block

  - **[Data]**: AnySize / Same Size - AES256 CTR - This is the encrypted version of the tar, containing the files

* The keys in this block have a different size when encrypted and decrypted
