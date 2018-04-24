package main

type Packet struct {
	channel chan []byte
	buf     []byte
}

func NewPacket(c chan []byte, buf []byte) *Packet {
	return &Packet{c, buf}
}

func (p *Packet) Broadcast() (n int) {
	for e := clientChannels.Front(); e != nil; e = e.Next() {
		channel := e.Value.(chan []byte)
		if channel != p.channel {
			channel <- p.buf
			n++
		}
	}
	trafficicity += len(p.buf)
	verbose("Broadcasted to", n, "clients")
	return
}
