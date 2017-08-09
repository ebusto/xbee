package xbee

import (
	"errors"
	"io"
	"log"
	"sync"
	"time"
)

type Node struct {
	Addr uint16
	Name string
	RSSI byte
}

type Radio struct {
	cf map[byte]FrameFn
	in map[uint16]chan Frame
	rw io.ReadWriter

	sync.Mutex
}

type (
	FrameFn func(Frame) bool
)

func NewRadio(rw io.ReadWriter) *Radio {
	r := &Radio{
		cf: make(map[byte]FrameFn),
		in: make(map[uint16]chan Frame),
		rw: rw,
	}

	go r.rx()

	return r
}

const (
	alarmDelay = time.Second * 5
)

// Address sets the radio's 16 bit address.
func (r *Radio) Address(addr uint16) error {
	return r.tx(nil, TypeAddress16(addr))
}

// Discover returns other nodes in the same PAN.
func (r *Radio) Discover() ([]Node, error) {
	var nodes []Node

	fn := func(f Frame) bool {
		b := f.Data()

		if len(b) == 0 {
			return true
		}

		na := b[0:2]           // 16-bit address.
		nn := b[11 : len(b)-1] // Identifier, ignore NULL.
		ns := b[10]            // RSSI.

		n := Node{Addr: Uint16(na), Name: string(nn), RSSI: ns}

		nodes = append(nodes, n)

		return false
	}

	return nodes, r.tx(fn, TypeDiscover())
}

// Identifier sets the radio's identifier.
func (r *Radio) Identifier(id string) error {
	return r.tx(nil, TypeIdentifier(id))
}

// TX sends the payload to the destination address.
func (r *Radio) TX(addr uint16, p []byte) error {
	return r.tx(nil, TypeTx16(addr), p)
}

// txStatus sends the payload and returns the status, ignoring any response.
func (r *Radio) tx(fn FrameFn, p ...[]byte) error {
	f := NewFrame(p...)

	r.Lock()

	// Send the frame to the radio.
	if err := Encode(r.rw, f); err != nil {
		return err
	}

	alarm := time.NewTimer(alarmDelay)
	frame := make(chan Frame)

	r.cf[f.Id()] = func(f Frame) bool {
		done := true

		if fn != nil {
			done = fn(f)
		}

		if done {
			alarm.Stop()
			frame <- f
		}

		return done
	}

	r.Unlock()

	var err error

	select {
	case <-alarm.C:
		err = errors.New("alarm")
	case f = <-frame:
		err = f.Status()
	}

	return err
}

func (r *Radio) RX(addr uint16) chan Frame {
	r.Lock()
	defer r.Unlock()

	r.in[addr] = make(chan Frame)

	return r.in[addr]
}

func (r *Radio) rx() {
	for {
		f, err := Decode(r.rw)

		if err != nil {
			log.Println(err)
			continue
		}

		r.Lock()

		// Dispatch the frame. Unhandled frames are silently discarded.
		switch f.Type() {
		case FrameTypeAtStatus, FrameTypeTxStatus:
			fn, ok := r.cf[f.Id()]

			if !ok {
				continue
			}

			if done := fn(f); done {
				delete(r.cf, f.Id())
			}

		case FrameTypeRx16:
			if ch, ok := r.in[f.Address16()]; ok {
				ch <- f
			}
		}

		r.Unlock()
	}
}
