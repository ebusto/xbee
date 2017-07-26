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
	bf   *bytes.Buffer
	err  error
	rd   *Radio

	sync.Mutex
}

func NewStream(rd *Radio, addr uint16) *Stream {
	s := &Stream{
		addr: addr,
		bf:   new(bytes.Buffer),
		rd:   rd,
	}

	go s.read()

	return s
}

func (s *Stream) Error() error {
	return s.err
}

func (s *Stream) Read(p []byte) (int, error) {
	s.Lock()

	n, err := s.bf.Read(p)

	s.Unlock()

	return n, err
}

func (s *Stream) read() {
	ch := s.rd.RX(s.addr)

	for f := range ch {
		s.Lock()

		_, s.err = s.bf.Write(f.Data())

		s.Unlock()
	}
}

func (s *Stream) Write(p []byte) (int, error) {
	n := 0

	write := func(b []byte) error {
		err := s.rd.TX(s.addr, b)

		if err == nil {
			n += len(b)
		}

		return err
	}

	for i := maxLen; len(p) > maxLen; i += maxLen {
		if err := write(p[:maxLen]); err != nil {
			return n, err
		}

		p = p[maxLen:]
	}

	if len(p) > 0 {
		return n, write(p)
	}

	return n, nil
}
