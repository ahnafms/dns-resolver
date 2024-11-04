package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

type DNSMessage struct {
	header   *Header
	question *Question
}

type Header struct {
	ID      uint16
	Flags   uint16
	QDCOUNT uint16
	ANCOUNT uint16
	NSCOUNT uint16
	ARCOUNT uint16
}

type Question struct {
	QNAME  string
	QTYPE  uint16
	QCLASS uint16
}

func NewDNSMessage(domain string) *DNSMessage {
	return &DNSMessage{
		header: &Header{
			ID:      22,
			Flags:   1 << 7,
			QDCOUNT: 1,
			ANCOUNT: 0,
			NSCOUNT: 0,
			ARCOUNT: 0,
		},
		question: &Question{
			QNAME:  domain,
			QTYPE:  1,
			QCLASS: 1,
		},
	}
}

func writeToBuff(buffer *bytes.Buffer, data any) error {
	return binary.Write(buffer, binary.BigEndian, data)
}

func domainToBuffer(domain string) []byte {
	parts := strings.Split(domain, ".")

	var domainBuf bytes.Buffer

	for _, part := range parts {
		if err := binary.Write(&domainBuf, binary.BigEndian, uint16(len(part))); err != nil {
			fmt.Println("Error writing length: ", err)
			return nil
		}
		if _, err := domainBuf.WriteString(part); err != nil {
			fmt.Println("Error writing string: ", err)
			return nil
		}
	}

	return domainBuf.Bytes()
}

func convertStructToBinary(dm *DNSMessage) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := writeToBuff(buf, dm.header); err != nil {
		return nil, fmt.Errorf("error writing ID: %v", err)
	}

	if err := writeToBuff(buf, domainToBuffer(dm.question.QNAME)); err != nil {
		return nil, fmt.Errorf("error writing QR: %v", err)
	}

	if err := writeToBuff(buf, uint16(dm.question.QTYPE)); err != nil {
		return nil, fmt.Errorf("error writing QR: %v", err)
	}

	if err := writeToBuff(buf, uint16(dm.question.QCLASS)); err != nil {
		return nil, fmt.Errorf("error writing QR: %v", err)
	}

	return buf.Bytes(), nil
}

func parseDNSMessage(message []byte) (*DNSMessage, error) {
	var header Header
	var question Question

	// header section

	if err := binary.Read(bytes.NewReader(message[:12]), binary.BigEndian, &header); err != nil {
		return nil, fmt.Errorf("error writing binary %v", err)
	}

	// question section
	start := 12
	var qnameBuf bytes.Buffer

	for {
		if message[start] == 0 {
			break
		}
		qnameBuf.WriteByte(message[start])

		start += 1
	}

	question.QNAME = qnameBuf.String()

	question.QTYPE = binary.BigEndian.Uint16(message[start : start+2])

	start += 2

	question.QCLASS = binary.BigEndian.Uint16(message[start : start+2])

	return &DNSMessage{header: &header, question: &question}, nil
}

func main() {
	message := NewDNSMessage("dns.google.com")
	test, err := convertStructToBinary(message)
	fmt.Println(hex.EncodeToString(test))
	if err != nil {
		fmt.Errorf("error convert struct to hex: %v", err)
		return
	}

	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println("Error creating udp connection", err)
		return
	}
	defer conn.Close()

	conn.Write(test)

	res := make([]byte, 512)
	_, err = conn.Read(res)
	if err != nil {
		fmt.Println("Error reading udp connection", err)
		return
	}

	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, res)

	responseId := binary.BigEndian.Uint16(res[:2])
	requestId := binary.BigEndian.Uint16(test[:2])
	fmt.Println(responseId, requestId, string(res))
}
