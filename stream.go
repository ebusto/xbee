package xbee

import (
	"bytes"
)

const (
	maxLen = 100 // Max payload. Xbee datasheet, page 62.
)

type Stream struct {
	ch chan Frame    // Incoming frames.
	pa uint16        // Peer address.
	rb *bytes.Buffer // Read buffer.
	rd *Radio        // Radio.
}

func NewStream(rd *Radio, addr uint16) *Stream {
	return &Stream{
		ch: rd.RX(addr),
		pa: addr,
		rb: new(bytes.Buffer),
		rd: rd,
	}
}

func (s *Stream) Read(p []byte) (int, error) {
	if s.rb.Len() > 0 {
		return s.rb.Read(p)
	}

	f := <-s.ch

	if _, err := s.rb.Write(f.Data()); err != nil {
		return 0, err
	}

	return s.rb.Read(p)
}

func (s *Stream) ReadByte() (byte, error) {
	var c [1]byte

	_, err := s.Read(c[:])

	return c[0], err
}

func (s *Stream) Write(p []byte) (int, error) {
	l := maxLen
	n := 0

	for i := l; len(p) > l; i += l {
		if err := s.write(p[:l], &n); err != nil {
			return n, err
		}

		p = p[l:]
	}

	return n, s.write(p, &n)
}

func (s *Stream) WriteByte(c byte) error {
	return s.write([]byte{c}, nil)
}

func (s *Stream) write(p []byte, n *int) error {
	if len(p) == 0 {
		return nil
	}

	err := s.rd.TX(s.pa, p)

	if err == nil && n != nil {
		*n += len(p)
	}

	return err
}
