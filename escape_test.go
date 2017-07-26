package xbee

import (
	"bytes"
	"testing"
)

var (
	testEscaped   = []byte{0x00, 0x7D, 0x31, 0x12, 0x7D, 0x33}
	testUnescaped = []byte{0x00, 0x11, 0x12, 0x13}
)

func TestEscapedReader(t *testing.T) {
	d := make([]byte, len(testUnescaped))
	r := &EscapedReader{bytes.NewBuffer(testEscaped)}

	if _, err := r.Read(d); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(d, testUnescaped) {
		t.Fatalf("Received %x, expected %x", d, testUnescaped)
	}
}

func TestEscapedWriter(t *testing.T) {
	b := new(bytes.Buffer)
	w := &EscapedWriter{b}

	if _, err := w.Write(testUnescaped); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(b.Bytes(), testEscaped) {
		t.Fatalf("Received %x, expected %x", b.Bytes(), testEscaped)
	}
}
