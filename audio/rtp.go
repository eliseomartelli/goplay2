package audio

import (
	"encoding/binary"
	aac "github.com/albanseurat/go-fdkaac"
	"github.com/brutella/hc/crypto/chacha20poly1305"
	"github.com/pion/rtp"
)

type TcpPacket struct {
	rtp.Packet
	SequenceNumber uint32
}

type PCMFrame struct {
	TcpPacket
	pcmData []byte
}

func (p *PCMFrame) Data() []byte {
	return p.pcmData
}

func NewFrame(aacDecoder *aac.AacDecoder, sharedKey []byte, rawPacket []byte) (*PCMFrame, error) {
	var err error
	packet := TcpPacket{}
	if err = packet.Unmarshal(rawPacket); err != nil {
		return nil, err
	}
	var seqBytes [4]byte
	copy(seqBytes[1:], rawPacket[1:4])
	packet.Marker = false  // used by apple in sequenceNumber
	packet.PayloadType = 0 // used by apple in sequenceNumber
	packet.SequenceNumber = binary.BigEndian.Uint32(seqBytes[:])
	message := packet.Payload[:len(packet.Payload)-24]
	nonce := packet.Payload[len(packet.Payload)-8:]
	var mac [16]byte
	copy(mac[:], packet.Payload[len(packet.Payload)-24:len(packet.Payload)-8])
	aad := packet.Raw[4:0xc]
	decrypted, err := chacha20poly1305.DecryptAndVerify(sharedKey, nonce, message, mac, aad)
	if err != nil {
		return nil, err
	}
	decode, err := aacDecoder.Decode(decrypted)
	if err != nil {
		return nil, err
	}
	return &PCMFrame{packet, decode}, nil
}
