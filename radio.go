package xbee

import (
	"errors"
	"io"
	"log"
	"sync"
	"time"
)

type Radio struct {
	cf map[byte]chan Frame
	in map[uint16]chan Frame
	lw time.Time
	rw io.ReadWriter

	sync.Mutex
}

func NewRadio(rw io.ReadWriter) *Radio {
	r := &Radio{
		cf: make(map[byte]chan Frame),
		in: make(map[uint16]chan Frame),
		lw: time.Now(),
		rw: rw,
	}

	go r.rx()

	return r
}

const (
	alarmDelay = time.Second * 2
	writeDelay = time.Millisecond * 10
)

// Address sets the radio's 16 bit address.
func (r *Radio) Address(addr uint16) error {
	return r.txStatus(TypeAddress16(addr))
}

// TX sends the payload to the destination address.
func (r *Radio) TX(addr uint16, p []byte) error {
	return r.txStatus(TypeTx16(addr), p)
}

// txStatus sends the payload and returns the status, ignoring any response.
func (r *Radio) txStatus(p ...[]byte) error {
	_, err := r.tx(p...)

	return err
}

// tx encapsulates the payload in a complete frame, writes it to the radio,
// and waits for confirmation.
func (r *Radio) tx(p ...[]byte) (Frame, error) {
	r.Lock()

	if el := time.Since(r.lw); el < writeDelay {
		time.Sleep(writeDelay - el)
	}

	r.lw = time.Now()

	f := NewFrame(p...)

	// Send the frame to the radio.
	if err := Encode(r.rw, f); err != nil {
		r.Unlock()

		return nil, err
	}

	// Listen for confirmation.
	ch := make(chan Frame)

	r.cf[f.Id()] = ch

	r.Unlock()

	alarm := time.NewTimer(alarmDelay)

	// Receive the response frame.
	select {
	case f = <-ch:
		alarm.Stop()

	case <-alarm.C:
		return nil, errors.New("alarm")
	}

	return f, f.Status()
}

func (r *Radio) RX(addr uint16) chan Frame {
	ch := make(chan Frame, 256)

	r.Lock()
	r.in[addr] = ch
	r.Unlock()

	return ch
}

func (r *Radio) rx() {
	for {
		f, err := Decode(r.rw)

		if err != nil {
			log.Println(err)
			continue
		}

		r.Lock()

		// Dispatch the frame. Unhandled frame types are silently discarded.
		switch f.Type() {
		case FrameTypeAtStatus, FrameTypeTxStatus:
			if ch, ok := r.cf[f.Id()]; ok {
				ch <- f
			}

			delete(r.cf, f.Id())

		case FrameTypeRx16:
			if ch, ok := r.in[f.Address16()]; ok {
				ch <- f
			}

		default:
			log.Printf("rx: unhandled: %x", f)
		}

		r.Unlock()
	}
}
