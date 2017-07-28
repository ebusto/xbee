package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/ebusto/xbee"
	"github.com/tarm/serial"
)

const (
	AddrLocal  = 0x3344
	AddrRemote = 0x1122
)

type Packet struct {
	Timestamp time.Time
	Payload   string
}

func main() {
	cn, err := serial.OpenPort(
		&serial.Config{Name: os.Args[1], Baud: 57600},
	)

	if err != nil {
		log.Fatal(err)
	}

	r := xbee.NewRadio(cn)
	s := xbee.NewStreamReader(r, AddrRemote)

	if err := r.Address(AddrLocal); err != nil {
		log.Printf("Unable to set local address: %s", err)
	}

	for {
		var p Packet

		d := json.NewDecoder(s)

		if err := d.Decode(&p); err != nil {
			log.Println(err)
		}

		log.Println(p)
	}
}
