package xbee

import (
	"encoding/hex"
	"testing"
)

func TestFrameAddress16(t *testing.T) {
	f := NewFrame(TypeAddress16(0x1122))

	t.Log(hex.EncodeToString(f))
}

func TestFrameAtCommand(t *testing.T) {
	f := NewFrame(TypeAtCommand(), []byte("MY"))

	t.Log(hex.EncodeToString(f))
}

func TestFrameTx16(t *testing.T) {
	f := NewFrame(TypeTx16(10), []byte("Felicia"))

	t.Log(hex.EncodeToString(f))
}

func TestFrameTx64(t *testing.T) {
	f := NewFrame(TypeTx64(16384), []byte("Felix"))

	t.Log(hex.EncodeToString(f))
}
