package xbee

import (
	"fmt"
	"io"
	"log"
	"sync"
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

// Address gets or sets the radio's 16 bit address.
func (r *Radio) Address(addr uint16) {
	f, err := r.tx(NewFrame(TypeAddress16(addr)))

	if err != nil {
		panic(err)
	}

	if f.Status() != 0x00 {
		panic("non-zero status")
	}
}

// TX sends the payload to the destination address.
func (r *Radio) TX(addr uint16, b []byte) error {
	f, err := r.tx(NewFrame(TypeTx16(addr), b))

	if err != nil {
		return err
	}

	if f.Status() != 0x00 {
		return fmt.Errorf("tx: status: %x", f.Status())
	}

	return nil
}

func (r *Radio) tx(f Frame) (Frame, error) {
	r.Lock()

	if err := Encode(r.rw, f); err != nil {
		r.Unlock()

		return nil, err
	}

	c := make(chan Frame)

	r.cf[f.Id()] = c
	r.Unlock()

	return <-c, nil
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
			log.Println(err)
		}

		r.Lock()

		// Dispatch the frame.
		switch f.Type() {
		case FrameTypeAtStatus, FrameTypeTxStatus:
			if c, ok := r.cf[f.Id()]; ok {
				c <- f
				delete(r.cf, f.Id())
			}

		case FrameTypeModemStatus:
			log.Printf("Modem: % X", f)

		case FrameTypeRx16:
			if c, ok := r.in[f.Address16()]; ok {
				c <- f
			}

		case FrameTypeRx64:
			log.Printf("Rx64: % X", f)

		default:
			log.Printf("Unexpected frame: %X", f)
		}

		r.Unlock()
	}
}
