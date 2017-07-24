package xbee

import (
	"bytes"
	"sync"
)

const (
	maxLen = 80
)

type Stream struct {
	addr uint16
	rb   *bytes.Buffer
	rd   *Radio

	sync.Mutex
}

func NewStream(rd *Radio, addr uint16) *Stream {
	rb := new(bytes.Buffer)

	s := &Stream{addr: addr, rb: rb, rd: rd}

	go s.read()

	return s
}

func (s *Stream) read() {
	for f := range s.rd.Recv(s.addr) {
		s.Lock()

		// TODO: Check errors.
		s.rb.Write(f.Data())

		s.Unlock()
	}
}

func (s *Stream) Read(p []byte) (int, error) {
	s.Lock()
	defer s.Unlock()

	return s.rb.Read(p)
}

func (s *Stream) Write(p []byte) (int, error) {
	d := [][]byte{}
	n := 0

	for i := maxLen; len(p) > maxLen; i += maxLen {
		d = append(d, p[:maxLen])
		p = p[maxLen:]
	}

	if len(p) > 0 {
		d = append(d, p)
	}

	for _, b := range d {
		if err := s.rd.Send(s.addr, b); err != nil {
			return n, err
		}

		n += len(b)
	}

	return n, nil
}
