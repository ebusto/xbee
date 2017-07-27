package xbee

import (
	"bytes"
)

type StreamReader struct {
	b bytes.Buffer
	c chan Frame
}

func NewStreamReader(rd *Radio, addr uint16) *StreamReader {
	return &StreamReader{bytes.Buffer{}, rd.RX(addr)}
}

func (r *StreamReader) Read(p []byte) (int, error) {
	if r.b.Len() > 0 {
		return r.b.Read(p)
	}

	f := <-r.c

	if _, err := r.b.Write(f.Data()); err != nil {
		return 0, err
	}

	return r.b.Read(p)
}

type StreamWriter struct {
	addr uint16
	rd   *Radio
}

func NewStreamWriter(rd *Radio, addr uint16) *StreamWriter {
	return &StreamWriter{addr, rd}
}

const (
	maxLen = 100 // Max payload. Xbee datasheet, page 62.
)

func (w *StreamWriter) Write(p []byte) (int, error) {
	l := maxLen
	n := 0

	for i := l; len(p) > l; i += l {
		if err := w.write(p[:l], &n); err != nil {
			return n, err
		}

		p = p[l:]
	}

	return n, w.write(p, &n)
}

func (w *StreamWriter) write(p []byte, n *int) error {
	if len(p) == 0 {
		return nil
	}

	err := w.rd.TX(w.addr, p)

	if err == nil {
		*n += len(p)
	}

	return err
}
