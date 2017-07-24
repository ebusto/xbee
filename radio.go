package xbee

import (
	"bufio"
	"io"
	"log"
	"sync"
)

type Radio struct {
	cf map[byte]chan Frame
	cn io.ReadWriter
	rx map[uint16]chan Frame

	sync.Mutex
}

func NewRadio(cn io.ReadWriter) *Radio {
	r := &Radio{
		cf: make(map[byte]chan Frame),
		cn: cn,
		rx: make(map[uint16]chan Frame),
	}

	go r.recv()

	return r
}

// Address gets or sets the radio's 16 bit address.
func (r *Radio) Address(addr ...uint16) uint16 {
	b := []byte("MY")

	if len(addr) > 0 {
		b = append(b, 0x00, 0x00)
		PutUint16(b[2:], addr[0])
	}

	f, err := r.tx(NewFrame(TypeAtCommand(), b))

	if err != nil {
		panic(err)
	}

	if len(f) > 9 {
		return Uint16(f[8:])
	}

	return 0
}

// Send transmits the payload to the destination address.
func (r *Radio) Send(addr uint16, p []byte) error {
	n := 0

	for n < 5 {
		f, err := r.tx(NewFrame(TypeTx16(addr), p))

		if err != nil {
			log.Fatalf("Tx16: %s", err)
		}

		if f.Status() == 0x00 {
			return nil
		}

		log.Printf("[%d] status: %X", n, f.Status())
		n++
	}

	return nil
}

func (r *Radio) Recv(addr uint16) chan Frame {
	ch := make(chan Frame)

	r.Lock()
	r.rx[addr] = ch
	r.Unlock()

	return ch
}

func (r *Radio) readBytes(br io.ByteReader, p []byte) error {
	e := false // Next byte escaped?
	n := 0     // Total bytes read.

	for n < len(p) {
		b, err := br.ReadByte()

		if err != nil {
			return err
		}

		// Next byte is escaped.
		if b == 0x7D {
			e = true
			continue
		}

		// This byte is escaped.
		if e {
			b = b ^ 0x20
			e = false
		}

		p[n] = b
		n++
	}

	return nil
}

var escape = map[byte]bool{
	0x11: true, // XON
	0x13: true, // XOFF
	0x7D: true, // Escape
	0x7E: true, // Start
}

func (r *Radio) frameWrite(f Frame) error {
	b := []byte{}

	// The first byte is not escaped.
	for i, c := range f {
		if escape[c] && i > 0 {
			b = append(b, 0x7D, c^0x20)
		} else {
			b = append(b, c)
		}
	}

	_, err := r.cn.Write(b)

	return err
}

func (r *Radio) recv() {
	br := bufio.NewReader(r.cn)

	for {
		f := make(Frame, 4)

		// Read the start byte.
		if err := r.readBytes(br, f[:1]); err != nil {
			panic(err)
		}

		// Is it the correct start byte?
		if f.Start() != 0x7E {
			log.Printf("Invalid start byte % X", f.Start())
			continue
		}

		// Read the length and type.
		if err := r.readBytes(br, f[1:]); err != nil {
			panic(err)
		}

		// Read the payload.
		p := make([]byte, f.Length())

		if err := r.readBytes(br, p); err != nil {
			panic(err)
		}

		// Append the payload to the frame.
		f = append(f, p...)

		// Is the checksum correct?
		if !f.Valid() {
			log.Fatal("Invalid frame: % X", f)
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
			if c, ok := r.rx[f.Address16()]; ok {
				c <- f
			}

		case FrameTypeRx64:
			log.Printf("Rx64: % X", f)

		default:
			log.Printf("Unknown frame type: % X", f.Type())
		}

		r.Unlock()
	}
}

func (r *Radio) tx(f Frame) (Frame, error) {
	r.Lock()

	if err := r.frameWrite(f); err != nil {
		r.Unlock()

		return nil, err
	}

	c := make(chan Frame)

	r.cf[f.Id()] = c
	r.Unlock()

	return <-c, nil
}
