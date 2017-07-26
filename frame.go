package xbee

import (
	"fmt"
	"io"
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

var seq = NewSequence()

type Frame []byte

func NewFrame(bs ...[]byte) Frame {
	f := Frame{
		0x7E, // 0: Start
		0x00, // 1: Length MSB
		0x00, // 2: Length LSB
	}

	for _, b := range bs {
		f = append(f, b...)
	}

	// Next sequence number.
	f[4] = <-seq

	l := f[1:] // Length
	p := f[3:] // Payload

	PutUint16(l, uint16(len(p)))

	return append(f, f.Checksum())
}

func TypeAddress16(addr uint16) []byte {
	b := TypeAtCommand()
	b = append(b, []byte("MY")...)
	b = append(b, 0x00, 0x00)

	PutUint16(b[4:], addr)

	return b
}

func TypeAtCommand() []byte {
	return []byte{
		0x08, // 3: Type
		0x00, // 4: ID
	}
}

func TypeTx16(addr uint16) []byte {
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

func TypeTx64(addr uint64) []byte {
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

func Decode(r io.Reader) (Frame, error) {
	f := make(Frame, 4)

	// Read the unescaped start byte.
	if _, err := r.Read(f[0:1]); err != nil {
		return nil, err
	}

	// Is it the correct start byte?
	if f.Start() != 0x7E {
		return nil, fmt.Errorf("invalid start byte: %x", f.Start())
	}

	r = &EscapedReader{r}

	// Read the length and type.
	if _, err := r.Read(f[1:]); err != nil {
		return nil, err
	}

	// Read the payload.
	p := make([]byte, f.Length())

	if _, err := r.Read(p); err != nil {
		return nil, err
	}

	// Append the payload to the frame.
	f = append(f, p...)

	// Is the checksum correct?
	if !f.Valid() {
		return nil, fmt.Errorf("invalid sum: %x", f.Sum())
	}

	return f, nil
}

func Encode(w io.Writer, f Frame) error {
	// Write the unescaped start byte.
	if _, err := w.Write(f[0:1]); err != nil {
		return err
	}

	// Write the escaped remainder.
	w = &EscapedWriter{w}

	_, err := w.Write(f[1:])

	return err
}
