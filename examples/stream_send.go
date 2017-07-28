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
	AddrLocal  = 0x1122
	AddrRemote = 0x3344
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
	s := xbee.NewStreamWriter(r, AddrRemote)

	if err := r.Address(AddrLocal); err != nil {
		log.Fatalf("Unable to set local address: %s", err)
	}

	for {
		p := Packet{time.Now(), "hello!"}
		e := json.NewEncoder(s)

		log.Println(p)

		if err := e.Encode(&p); err != nil {
			log.Println(err)
		}
	}
}
