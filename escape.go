package xbee

import (
	"io"
)

type EscapedReader struct {
	R io.Reader
}

type EscapedWriter struct {
	W io.Writer
}

func (r *EscapedReader) Read(b []byte) (int, error) {
	p := make([]byte, 1) // Single byte buffer.
	e := false           // Next byte escaped?
	n := 0               // Total bytes read.

	for n < len(b) {
		if _, err := r.R.Read(p); err != nil {
			return n, err
		}

		c := p[0]

		switch {
		case e == true:
			c = c ^ 0x20
			e = false

		case c == 0x7D:
			e = true
			continue
		}

		b[n] = c
		n++
	}

	return n, nil
}

func (w *EscapedWriter) Write(b []byte) (int, error) {
	escape := map[byte]bool{
		0x11: true, // XON
		0x13: true, // XOFF
		0x7D: true, // Escape
		0x7E: true, // Start
	}

	var e []byte

	for _, c := range b {
		if escape[c] {
			e = append(e, 0x7D, c^0x20)
		} else {
			e = append(e, c)
		}
	}

	return w.W.Write(e)
}
