package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
)

const BUFFER_SIZE int = 4096

type writer struct {
	buffer []byte
	off    int
	out    io.Writer
	ctr    cipher.Stream
}
type reader struct {
	encr []byte // Encrypted buffer
	decr []byte // Decrypted buffer
	in   io.Reader
	ctr  cipher.Stream
}

type AESWriter interface {
	Write(data []byte) (int, error)
	Flush()
}
type AESReader interface {
	Read(data []byte) (int, error)
}

func NewAesWriter(aesKey, iv []byte, out io.Writer) (AESWriter, error) {
	aes, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	return &writer{
		buffer: make([]byte, 0, BUFFER_SIZE),
		off:    0,
		out:    out,
		ctr:    cipher.NewCTR(aes, iv),
	}, nil
}

func NewAesReader(aesKey, iv []byte, in io.Reader) (AESReader, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	r := reader{
		decr: make([]byte, 0, BUFFER_SIZE),
		encr: make([]byte, BUFFER_SIZE),
		in:   in,
		ctr:  cipher.NewCTR(block, iv),
	}

	// Fill the encrypted buffer
	n, err := in.Read(r.encr)
	if err != nil {
		return nil, err
	}
	r.encr = r.encr[:n]

	return &r, nil
}

func (p *reader) addDecr(data, res []byte) []byte {
	eo := min(len(p.decr), len(data)-len(res))
	t := p.decr[:eo]

	d := make([]byte, len(p.decr)-eo, BUFFER_SIZE)
	copy(d, p.decr[eo:])
	p.decr = d

	return t
}

func (p *reader) Read(data []byte) (int, error) {
	if len(p.encr) == 0 {
		return 0, io.EOF
	}

	// Already "cached"
	if len(data) < len(p.decr) {
		n := copy(data, p.decr[:len(data)])
		p.decr = p.decr[len(data):]
		return n, nil
	}

	res := make([]byte, 0, max(BUFFER_SIZE, len(data))) // It would have been increased to BUFFER_SIZE anyways...
	for {

		// Append what can be appendend
		res = append(res, p.addDecr(data, res)...)
		if len(res) == len(data) {
			break
		}

		// If not enough, check if there's still more to decrypt
		if len(p.encr) == 0 {
			break
		}

		// Decrypt another chunk of data
		p.decr = p.decr[:len(p.encr)]
		p.ctr.XORKeyStream(p.decr, p.encr)

		// Refill the decrypted part
		n, err := p.in.Read(p.encr)
		if err != nil && err != io.EOF {
			return 0, err
		}
		p.encr = p.encr[:n]
	}

	n := copy(data, res)
	return n, nil
}

func (p *writer) Write(data []byte) (int, error) {
	n := 0
	for {
		eo := min(len(data), BUFFER_SIZE-len(p.buffer))
		n += eo
		if eo == 0 {
			return n, nil
		}

		p.buffer = append(p.buffer, data[:eo]...)
		data = data[eo:]
		if len(p.buffer) == BUFFER_SIZE {
			p.Flush()
		}
	}
}
func (p *writer) Flush() {
	n := len(p.buffer)
	outBuf := make([]byte, n)
	p.ctr.XORKeyStream(outBuf, p.buffer)
	p.out.Write(outBuf)
	p.buffer = p.buffer[:0]
}
