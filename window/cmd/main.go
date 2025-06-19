package main

import (
	"encoding/binary"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("unix", "/tmp/gocomp.sock")
	if err != nil {
		log.Fatal("cannot connect:", err)
	}
	defer conn.Close()

	buf := make([]byte, 13)
	buf[0] = 0x01
	buf[1] = 0x00
	binary.LittleEndian.PutUint16(buf[2:], 50)  // x
	binary.LittleEndian.PutUint16(buf[4:], 50)  // y
	binary.LittleEndian.PutUint16(buf[6:], 100) // width
	binary.LittleEndian.PutUint16(buf[8:], 80)  // height
	buf[10] = 255                               // R
	buf[11] = 0                                 // G
	buf[12] = 0                                 // B

	conn.Write(buf)
}
