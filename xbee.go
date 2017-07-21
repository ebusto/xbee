package xbee

import (
	"encoding/binary"
)

type Frame []byte

var (
	PutUint16 = binary.BigEndian.PutUint16
	PutUint64 = binary.BigEndian.PutUint64
)

func NewFrame(b []byte) Frame {
	f := []byte{
		0x7E, // 0: Delimiter
		0x00, // 1: Length MSB
		0x00, // 2: Length LSB
	}

	f = append(f, b...)

	l := f[1:] // Length
	p := f[3:] // Payload

	PutUint16(l, uint16(len(p)))

	return append(f, Checksum(p))
}

func AtCommand(id byte, cmd string) []byte {
	return []byte{
		0x08,   // 3: Type
		id,     // 4: ID
		cmd[0], // 5: Command(0)
		cmd[1], // 6: Command(1)
	}
}

func Tx16(id byte, addr uint16, data []byte) []byte {
	b := []byte{
		0x01, // 3: Type
		id,   // 4: ID
		0x00, // 5: Address MSB
		0x00, // 6: Address LSB
		0x00, // 7: Options
	}

	PutUint16(b[2:], addr)

	return append(b, data...)
}

func Tx64(id byte, addr uint64, data []byte) []byte {
	b := []byte{
		0x00, //  3: Type
		id,   //  4: ID
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

	return append(b, data...)
}

func Checksum(b []byte) byte {
	var sum byte

	for _, c := range b {
		sum += c
	}

	return 0xFF - sum
}

func Escape(b []byte) []byte {
	escapeBytes := map[byte]bool{
		0x11: true, // XON
		0x13: true, // XOFF
		0x7D: true, // Escape
		0x7E: true, // Frame start
	}

	var e []byte

	for i, c := range b {
		switch escapeBytes[c] && i > 0 {
		case true:
			e = append(e, 0x7D, c^0x20)
		case false:
			e = append(e, c)
		}
	}

	return e
}
