package xbee

import (
	"bytes"
	"testing"
)

func TestEscapedReader(t *testing.T) {
	esc := []byte{0x00, 0x7D, 0x31, 0x12, 0x7D, 0x33}
	exp := []byte{0x00, 0x11, 0x12, 0x13}

	d := make([]byte, 4)

	r := &EscapedReader{bytes.NewBuffer(esc)}

	if _, err := r.Read(d); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(d, exp) {
		t.Fatalf("Received %x, expected %x", d, exp)
	}
}
