package main

import (
	"log"
	"os"
	"time"

	"github.com/ebusto/xbee"
	"github.com/tarm/serial"
)

const (
	AddrLocal  = 0x1122
	AddrRemote = 0x3344
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

	for {
		b, err := time.Now().MarshalText()

		if err != nil {
			log.Fatal(err)
		}

		if err := r.TX(AddrRemote, b); err != nil {
			log.Println(err)
		}
	}
}
