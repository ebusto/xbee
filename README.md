The xbee package is for using Series 1 XBee devices in API mode to send and receive frames.

# Example
```go
package main

import (
	"net"

	"github.com/tarm/serial"
	"github.com/ebusto/xbee"
)

func main() {
	cn, err := serial.OpenPort(
	                &serial.Config{Name: os.Args[1], Baud: 57600},
	        )
	
	        if err != nil {
	                log.Fatal(err)
	        }
	
	        r.Address(0x1122)) // Set the 16-bit local address.

	        r := xbee.NewRadio(cn)
	        s := xbee.NewStream(r, 0x3344) // Stream to 16-bit destination address.
	
	
	        for {
	                b, err := time.Now().MarshalText()
	
	                if err != nil {
	                        log.Fatal(err)
	                }
	
	                if _, err := s.Write(b); err != nil {
	                        log.Fatal(err)
	                }
	        }
	}
}
```
