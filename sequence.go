package xbee

func NewSequence() chan byte {
	c := make(chan byte)

	go func() {
		for i := byte(1); ; i++ {
			if i > 0 {
				c <- i
			}
		}
	}()

	return c
}
