package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/ebusto/xbee"
	"github.com/tarm/serial"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <radio> <id>", os.Args[0])
	}

	cn, err := serial.OpenPort(
		&serial.Config{Name: os.Args[1], Baud: 57600},
	)

	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().Unix())

	addr := uint16(rand.Int())

	r := xbee.NewRadio(cn)

	log.Printf("Setting address to: %X", addr)

	if err := r.Address(addr); err != nil {
		log.Fatalf("Unable to set local address: %s", err)
	}

	log.Printf("Setting identifier to: %s", os.Args[2])

	if err := r.Identifier(os.Args[2]); err != nil {
		log.Fatalf("Unable to set identifier: %s", err)
	}

	for {
		time.Sleep(time.Second)

		nodes, err := r.Discover()

		if err != nil {
			log.Fatalf("Unable to discover nodes: %s", err)
		}

		log.Printf("%#v", nodes)
	}
}
