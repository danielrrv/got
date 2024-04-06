package internal

type Packet struct {
	buff []byte
	len int
}

func AllocatePacket(length int) *Packet {
	return &Packet{
		buff: make([]byte, 0,length),
		len: 0,
	}
}

func (p *Packet) Set(data ...[]byte) {
	for _, d := range data {
		p.buff = append(p.buff, d...)
		p.len += len(d)
	}
}
