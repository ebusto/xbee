package xbee

import (
	"io"
	"log"
	"sync"
	"time"
)

type Radio struct {
	cf map[byte]chan Frame
	in map[uint16]chan Frame
	rw io.ReadWriter

	sync.Mutex
}

func NewRadio(rw io.ReadWriter) *Radio {
	r := &Radio{
		cf: make(map[byte]chan Frame),
		in: make(map[uint16]chan Frame),
		rw: rw,
	}

	go r.rx()

	return r
}

// Address sets the radio's 16 bit address.
func (r *Radio) Address(addr uint16) error {
	return r.tx(TypeAddress16(addr))
}

// TX sends the payload to the destination address.
func (r *Radio) TX(addr uint16, p []byte) error {
	return r.tx(TypeTx16(addr), p)
}

// tx encapsulates the payload in a complete frame, writes it to the radio,
// and waits for confirmation.
func (r *Radio) tx(p ...[]byte) error {
	f := NewFrame(p...)

	r.Lock()

	// Send the frame to the radio.
	if err := Encode(r.rw, f); err != nil {
		r.Unlock()

		return err
	}

	// Listen for confirmation.
	ch := make(chan Frame)

	r.cf[f.Id()] = ch

	r.Unlock()

	// Receive the response frame.
	f = <-ch

	if err := f.Status(); err != nil {
		time.Sleep(time.Millisecond * 250)

		return err
	}

	return nil
}

func (r *Radio) RX(addr uint16) chan Frame {
	ch := make(chan Frame)

	r.Lock()
	r.in[addr] = ch
	r.Unlock()

	return ch
}

func (r *Radio) rx() {
	for {
		f, err := Decode(r.rw)

		if err != nil {
			log.Printf("rx: decode: %s", err)
			continue
		}

		r.Lock()

		// Dispatch the frame. Unhandled frame types are silently discarded.
		switch f.Type() {
		case FrameTypeAtStatus, FrameTypeTxStatus:
			if c, ok := r.cf[f.Id()]; ok {
				c <- f
				delete(r.cf, f.Id())
			}

		case FrameTypeRx16:
			if c, ok := r.in[f.Address16()]; ok {
				c <- f
			}
		}

		r.Unlock()
	}
}
