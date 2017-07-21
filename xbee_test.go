package xbee

import (
	"encoding/hex"
	"testing"
)

func TestFrameAtCommand(t *testing.T) {
	f := NewFrame(AtCommand(8, "MY"))

	t.Log(hex.EncodeToString(f))
}

func TestFrameTx16(t *testing.T) {
	f := NewFrame(Tx16(8, 10, []byte("Felicia")))

	t.Log(hex.EncodeToString(f))
}

func TestFrameTx64(t *testing.T) {
	f := NewFrame(Tx64(8, 16384, []byte("Felix")))

	t.Log(hex.EncodeToString(f))
}
