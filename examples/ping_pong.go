package main

import (
	"encoding/binary"
	"log"
	"math/rand"
	"os"

	"github.com/ebusto/xbee"
	"github.com/tarm/serial"
)

var players = map[string]uint16{
	"A": 0x1122,
	"B": 0x3344,
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <radio> A|B", os.Args[0])
	}

	cn, err := serial.OpenPort(
		&serial.Config{Name: os.Args[1], Baud: 57600},
	)

	if err != nil {
		log.Fatal(err)
	}

	var addrLocal, addrRemote uint16

	for name, addr := range players {
		if name == os.Args[2] {
			addrLocal = addr
		} else {
			addrRemote = addr
		}
	}

	r := xbee.NewRadio(cn)
	s := xbee.NewStream(r, addrRemote)

	if err := r.Address(addrLocal); err != nil {
		log.Fatalf("Unable to set local address: %s", err)
	}

	rand.Seed(int64(addrLocal))

	log.Printf("%s: %x -> %x", os.Args[2], addrLocal, addrRemote)

	buf := make([]byte, binary.MaxVarintLen64)

	v := rand.Int63n(int64(addrLocal))
	n := binary.PutVarint(buf, v)

	if _, err := s.Write(buf[:n]); err != nil {
		log.Fatal(err)
	}

	for {
		v, err := binary.ReadVarint(s)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("read %d, sending %d", v, v+1)

		n := binary.PutVarint(buf, v+1)

		for {
			if _, err := s.Write(buf[:n]); err != nil {
				log.Println(err)
				continue
			}

			break
		}
	}
}
