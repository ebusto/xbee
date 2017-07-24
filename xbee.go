package xbee

import (
	"encoding/binary"
)

var (
	PutUint16 = binary.BigEndian.PutUint16
	PutUint64 = binary.BigEndian.PutUint64
	Uint16    = binary.BigEndian.Uint16
	Uint64    = binary.BigEndian.Uint64
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
