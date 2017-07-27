package main

import (
	"log"
	"os"

	"github.com/ebusto/xbee"
	"github.com/tarm/serial"
)

const (
	AddrLocal  = 0x3344
	AddrRemote = 0x1122
)

func main() {
	cn, err := serial.OpenPort(
		&serial.Config{Name: os.Args[1], Baud: 57600},
	)

	if err != nil {
		log.Fatal(err)
	}

	r := xbee.NewRadio(cn)

	if err := r.Address(AddrLocal); err != nil {
		log.Fatalf("Unable to set local address: %s", err)
	}

	ch := r.RX(AddrRemote)

	for f := range ch {
		log.Printf("[%d | %d]: %s", f.Address16(), f.RSSI(), f.Data())
	}
}
