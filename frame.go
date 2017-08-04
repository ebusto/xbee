package xbee

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	FrameOffsetAddress  = 4
	FrameOffsetAtStatus = 7
	FrameOffsetData     = 8
	FrameOffsetId       = 4
	FrameOffsetLength   = 1
	FrameOffsetRSSI     = 6
	FrameOffsetStart    = 0
	FrameOffsetTxStatus = 5
	FrameOffsetType     = 3
)

var (
	PutUint16 = binary.BigEndian.PutUint16
	PutUint64 = binary.BigEndian.PutUint64
	Uint16    = binary.BigEndian.Uint16
	Uint64    = binary.BigEndian.Uint64
)

var seq = NewSequence()

type Frame []byte

// NewFrame returns a new frame with the specified payload.
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

// TypeAddress16 returns the payload for a 16-bit address assignment.
func TypeAddress16(addr uint16) []byte {
	b := TypeAtCommand()

	b = append(b, []byte("MY")...)
	b = append(b, 0x00, 0x00)

	PutUint16(b[4:], addr)

	return b
}

// TypeAtCommand returns the payload for an AT command request.
func TypeAtCommand() []byte {
	return []byte{
		0x08, // 3: Type
		0x00, // 4: ID
	}
}

// TypeDiscovery returns the payload for a node discovery request.
func TypeDiscover() []byte {
	b := TypeAtCommand()

	b = append(b, []byte("ND")...)

	return b
}

// TypeIdentifier returns the payload for a node identifier assignment.
func TypeIdentifier(id string) []byte {
	b := TypeAtCommand()

	b = append(b, []byte("NI")...)
	b = append(b, []byte(id)...)

	return b
}

// TypeTx16 returns the payload for transmitting to a 16-bit address.
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

// TypeTx64 returns the payload for transmitting to a 64-bit address.
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

// Address16 returns the 16-bit source address.
func (f Frame) Address16() uint16 {
	return Uint16(f[FrameOffsetAddress:])
}

// Address16 returns the 64-bit source address.
func (f Frame) Address64() uint64 {
	return Uint64(f[FrameOffsetAddress:])
}

// Checksum returns the calculated checksum.
func (f Frame) Checksum() byte {
	return 0xFF - f.Sum()
}

// Data returns the remainder of the frame, minus the checksum.
func (f Frame) Data() []byte {
	return f[FrameOffsetData : len(f)-1]
}

// Id returns the confirmation ID.
func (f Frame) Id() byte {
	return f[FrameOffsetId]
}

// Length returns the length.
func (f Frame) Length() int {
	return int(Uint16(f[FrameOffsetLength:]))
}

// RSSI returns the signal strength.
func (f Frame) RSSI() int {
	return int(f[FrameOffsetRSSI])
}

// Start returns the start byte.
func (f Frame) Start() byte {
	return f[FrameOffsetStart]
}

// Status returns the type specific status.
func (f Frame) Status() error {
	switch f.Type() {
	// XBee datasheet, page 60.
	case FrameTypeAtStatus:
		switch f[FrameOffsetAtStatus] {
		case 0x00:
			return nil
		case 0x01:
			return errors.New("unknown")
		case 0x02:
			return errors.New("invalid command")
		case 0x03:
			return errors.New("invalid parameter")
		}

	// Xbee datasheet, page 63.
	case FrameTypeTxStatus:
		switch f[FrameOffsetTxStatus] {
		case 0x00:
			return nil
		case 0x01:
			return errors.New("no ack received")
		case 0x02:
			return errors.New("cca failure")
		case 0x03:
			return errors.New("purged")
		}
	}

	return nil
}

// Sum returns the sum of all bytes, minus the start and length.
func (f Frame) Sum() byte {
	var sum byte

	for _, c := range f[FrameOffsetType:] {
		sum += c
	}

	return sum
}

// Type returns the type.
func (f Frame) Type() byte {
	return f[FrameOffsetType]
}

// Valid returns whether or not the frame is valid, that is, the sum of all
// bytes matches the expected value.
func (f Frame) Valid() bool {
	return 0xFF == f.Sum()
}

// Decode reads a complete frame from an io.Reader.
func Decode(r io.Reader) (Frame, error) {
	f := make(Frame, 4)

	for {
		// Read the unescaped start byte.
		if _, err := r.Read(f[:1]); err != nil {
			return nil, err
		}

		// Is it the correct start byte?
		if f.Start() == 0x7E {
			break
		}
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

// Encode writes a complete frame to an io.Writer.
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
