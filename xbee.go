package xbee

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"sync"
)

var (
	PutUint16 = binary.BigEndian.PutUint16
	PutUint64 = binary.BigEndian.PutUint64
	Uint16    = binary.BigEndian.Uint16
	Uint64    = binary.BigEndian.Uint64
)

const (
	FrameOffsetAddress  = 4
	FrameOffsetAtStatus = 8
	FrameOffsetData     = 8
	FrameOffsetId       = 4
	FrameOffsetLength   = 1
	FrameOffsetRSSI     = 6
	FrameOffsetStart    = 0
	FrameOffsetTxStatus = 5
	FrameOffsetType     = 3
)

const (
	FrameTypeAtCommand   = 0x08
	FrameTypeAtStatus    = 0x88
	FrameTypeModemStatus = 0x8A
	FrameTypeRx16        = 0x81
	FrameTypeRx64        = 0x80
	FrameTypeTx16        = 0x01
	FrameTypeTx64        = 0x00
	FrameTypeTxStatus    = 0x89
)

// Bytes that need to be escaped.
var escape = map[byte]bool{
	0x11: true, // XON
	0x13: true, // XOFF
	0x7D: true, // Escape
	0x7E: true, // Start
}

type Frame []byte

type XBee struct {
	cf map[byte]chan Frame
	cn io.ReadWriter
	rx map[uint16]chan Frame
	sn chan byte

	sync.Mutex
}

func New(cn io.ReadWriter) *XBee {
	x := &XBee{
		cf: make(map[byte]chan Frame),
		cn: cn,
		rx: make(map[uint16]chan Frame),
		sn: Sequence(),
	}

	go x.recv()

	return x
}

// Address gets or sets the radio's 16 bit address.
func (x *XBee) Address(addr ...uint16) uint16 {
	b := []byte("MY")

	if len(addr) > 0 {
		b = append(b, 0x00, 0x00)
		PutUint16(b[2:], addr[0])
	}

	r, err := x.tx(x.NewFrame(x.TypeAtCommand(), b))

	if err != nil {
		panic(err)
	}

	if len(r) > 9 {
		return Uint16(r[8:])
	}

	return 0
}

// Send transmits the payload to the destination address.
func (x *XBee) Send(addr uint16, p []byte) error {
	n := 0

	for n < 5 {
		f := x.NewFrame(x.TypeTx16(addr), p)

		r, err := x.tx(f)

		if err != nil {
			log.Fatalf("Tx16: %s", err)
		}

		if r.Status() == 0x00 {
			return nil
		}

		log.Printf("[%d] status: %X", n, r.Status())
		n++
	}

	return nil
}

func (x *XBee) Recv(addr uint16) chan Frame {
	ch := make(chan Frame)

	x.Lock()
	x.rx[addr] = ch
	x.Unlock()

	return ch
}

func (x *XBee) readBytes(r io.ByteReader, p []byte) error {
	e := false // Next byte escaped?
	n := 0     // Total bytes read.

	for n < len(p) {
		b, err := r.ReadByte()

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

func (x *XBee) recv() {
	br := bufio.NewReader(x.cn)

	for {
		f := make(Frame, 4)

		// Read the start byte.
		if err := x.readBytes(br, f[:1]); err != nil {
			panic(err)
		}

		// Is it the correct start byte?
		if f.Start() != 0x7E {
			log.Printf("Invalid start byte % X", f.Start())
			continue
		}

		// Read the length and type.
		if err := x.readBytes(br, f[1:]); err != nil {
			panic(err)
		}

		// Read the payload.
		p := make([]byte, f.Length())

		if err := x.readBytes(br, p); err != nil {
			panic(err)
		}

		// Append the payload to the frame.
		f = append(f, p...)

		// Is the checksum correct?
		if !f.Valid() {
			log.Fatal("Invalid frame: % X", f)
		}

		x.Lock()

		// Dispatch the frame.
		switch f.Type() {
		case FrameTypeAtStatus, FrameTypeTxStatus:
			if c, ok := x.cf[f.Id()]; ok {
				c <- f
				delete(x.cf, f.Id())
			}

		case FrameTypeModemStatus:
			log.Printf("Modem: % X", f)

		case FrameTypeRx16:
			if c, ok := x.rx[f.Address16()]; ok {
				c <- f
			}

		case FrameTypeRx64:
			log.Printf("Rx64: % X", f)

		default:
			log.Printf("Unknown frame type: % X", f.Type())
		}

		x.Unlock()
	}
}

func (x *XBee) tx(f Frame) (Frame, error) {
	x.Lock()

	if _, err := x.cn.Write(f.Escape()); err != nil {
		x.Unlock()

		return nil, err
	}

	c := make(chan Frame)

	x.cf[f.Id()] = c
	x.Unlock()

	return <-c, nil
}

func (x *XBee) NewFrame(bs ...[]byte) Frame {
	f := Frame{
		0x7E, // 0: Start
		0x00, // 1: Length MSB
		0x00, // 2: Length LSB
	}

	for _, b := range bs {
		f = append(f, b...)
	}

	// Next sequence number.
	f[4] = <-x.sn

	l := f[1:] // Length
	p := f[3:] // Payload

	PutUint16(l, uint16(len(p)))

	return append(f, f.Checksum())
}

func (x *XBee) TypeAtCommand() []byte {
	return []byte{
		0x08, // 3: Type
		0x00, // 4: ID
	}
}

func (x *XBee) TypeTx16(addr uint16) []byte {
	b := []byte{
		0x01, // 3: Type
		0x00, // 4: ID
		0x00, // 5: Address MSB
		0x00, // 6: Address LSB
		0x00, // 7: Options
	}

	PutUint16(b[2:], addr)

	return b
}

func (x *XBee) TypeTx64(addr uint64) []byte {
	b := []byte{
		0x00, //  3: Type
		0x00, //  4: ID
		0x00, //  5: Address MSB
		0x00, //  6: Address MSB
		0x00, //  7: Address MSB
		0x00, //  8: Address MSB
		0x00, //  9: Address LSB
		0x00, // 10: Address LSB
		0x00, // 11: Address LSB
		0x00, // 12: Address LSB
		0x00, // 13: Options
	}

	PutUint64(b[2:], addr)

	return b
}

func (f Frame) Address16() uint16 {
	return Uint16(f[FrameOffsetAddress:])
}

func (f Frame) Address64() uint64 {
	return Uint64(f[FrameOffsetAddress:])
}

func (f Frame) Checksum() byte {
	return 0xFF - f.Sum()
}

func (f Frame) Data() []byte {
	// Do not include the checksum.
	return f[FrameOffsetData : len(f)-1]
}

func (f Frame) Escape() []byte {
	b := []byte{0x7E}

	// The first byte is not escaped, only length onwards.
	for _, c := range f[FrameOffsetLength:] {
		if escape[c] {
			b = append(b, 0x7D, c^0x20)
		} else {
			b = append(b, c)
		}
	}

	return b
}

func (f Frame) Id() byte {
	return f[FrameOffsetId]
}

func (f Frame) Length() int {
	return int(Uint16(f[FrameOffsetLength:]))
}

func (f Frame) RSSI() int {
	return int(f[FrameOffsetRSSI])
}

func (f Frame) Start() byte {
	return f[FrameOffsetStart]
}

func (f Frame) Status() byte {
	switch f.Type() {
	case FrameTypeAtStatus:
		return f[FrameOffsetAtStatus]

	case FrameTypeTxStatus:
		return f[FrameOffsetTxStatus]
	}

	return 0xFF
}

func (f Frame) Sum() byte {
	var sum byte

	// The start and length bytes are not included.
	for _, c := range f[FrameOffsetType:] {
		sum += c
	}

	return sum
}

func (f Frame) Type() byte {
	return f[FrameOffsetType]
}

func (f Frame) Valid() bool {
	return 0xFF == f.Sum()
}

func Sequence() chan byte {
	c := make(chan byte)

	go func() {
		for i := byte(1); ; i++ {
			if i > 0 {
				c <- i
			}
		}
	}()

	return c
}
