package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"time"
)

func openConnection(host string, port uint16) net.Conn {
	dialer := &net.Dialer{Timeout: 2 * time.Second}
	connection, err := dialer.Dial("tcp", host+":"+strconv.Itoa(int(port)))
	if err != nil {
		panic("Could not open connection: " + err.Error())
	}
	return connection
}

func queryDevice(connection net.Conn, request string) []byte {
	scrambledText := scramble([]byte(request))
	bytesWritten, err := fmt.Fprint(connection, string(scrambledText))
	if err != nil || bytesWritten != len(scrambledText) {
		panic("Could not write command to connection")
	}
	err = connection.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		panic("Could not set read timeout on connection: " + err.Error())
	}
	buffer := make([]byte, 1024)
	bytesRead, err := bufio.NewReader(connection).Read(buffer)
	if err != nil {
		panic("Could not read from connection: " + err.Error())
	}
	return unscramble(buffer[:minInt(bytesRead, len(buffer))])
}

func minInt(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func scramble(b []byte) []byte {
	var iv byte = 171
	buffer := make([]byte, 4+len(b))

	writeUInt32ToBufferBigEndian(buffer, uint32(len(b)))
	for i, ch := range b {
		iv = byte(iv ^ ch)
		buffer[i+4] = iv
	}
	return buffer
}

func unscramble(b []byte) []byte {
	var iv byte = 171
	buffer := make([]byte, len(b)-4)

	expectedSize := int(b[3]) + int(b[2])<<8 + int(b[1])<<16 + int(b[0])<<24
	if expectedSize != len(b)-4 {
		panic("Unexpected reply size - expected " + strconv.Itoa(expectedSize) +
			" bytes but received " + strconv.Itoa(len(b)-4) + " bytes")
	}
	for i, ch := range b[4:] {
		buffer[i] = byte(iv ^ ch)
		iv = ch
	}
	return buffer
}

func writeUInt32ToBufferBigEndian(b []byte, i uint32) {
	b[0] = byte((i >> 24) & 0xff)
	b[1] = byte((i >> 16) & 0xff)
	b[2] = byte((i >> 8) & 0xff)
	b[3] = byte(i & 0xff)
}
