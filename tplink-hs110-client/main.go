package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

var commands = map[string]map[string]string{
	"system": {
		"get_sysinfo": `{"system":{"get_sysinfo":null}}`,
	},
	"emeter": {
		"get_realtime":    `{"emeter":{"get_realtime":{}}}`,
		"get_vgain_igain": `{"emeter":{"get_vgain_igain":{}}}`,
	},
}

func main() {
	if len(os.Args) != 3 {
		panic("This app needs exactly two arguments - IP/hostname and port")
	}
	host := os.Args[1]
	port, portErr := strconv.ParseUint(os.Args[2], 10, 16)
	if portErr != nil {
		panic("Port must be a number")
	}

	connection := OpenConnection(host, uint16(port))
	defer func() { _ = connection.Close() }()

	info := queryDevice(connection, commands["system"]["get_sysinfo"])
	var infoJson map[string]map[string]map[string]interface{}
	if err := json.Unmarshal(info, &infoJson); err != nil {
		panic("Could not unmarshal info JSON string: " + err.Error())
	}
	alias := infoJson["system"]["get_sysinfo"]["alias"].(string)
	id := infoJson["system"]["get_sysinfo"]["deviceId"].(string)
	mac := infoJson["system"]["get_sysinfo"]["mac"].(string)
	state := int(infoJson["system"]["get_sysinfo"]["relay_state"].(float64))
	onTime := int(infoJson["system"]["get_sysinfo"]["on_time"].(float64))

	realTime := queryDevice(connection, commands["emeter"]["get_realtime"])
	var realtimeJson map[string]map[string]map[string]float64
	if err := json.Unmarshal(realTime, &realtimeJson); err != nil {
		panic("Could not unmarshal real-time JSON string: " + err.Error())
	}
	voltageMv := int(realtimeJson["emeter"]["get_realtime"]["voltage_mv"])
	currentMa := int(realtimeJson["emeter"]["get_realtime"]["current_ma"])
	powerMw := int(realtimeJson["emeter"]["get_realtime"]["power_mw"])
	totalWh := int(realtimeJson["emeter"]["get_realtime"]["total_wh"])

	tags := `alias="` + alias + `",id="` + id + `",mac="` + mac + `"`

	fmt.Println("# TYPE state gauge")
	fmt.Println("state{" + tags + "} " + strconv.Itoa(state))
	fmt.Println("# TYPE on_time gauge")
	fmt.Println("on_time{" + tags + "} " + strconv.Itoa(onTime))
	fmt.Println("# TYPE voltage_mv gauge")
	fmt.Println("voltage_mv{" + tags + "} " + strconv.Itoa(voltageMv))
	fmt.Println("# TYPE current_ma gauge")
	fmt.Println("current_ma{" + tags + "} " + strconv.Itoa(currentMa))
	fmt.Println("# TYPE power_mw gauge")
	fmt.Println("power_mw{" + tags + "} " + strconv.Itoa(powerMw))
	fmt.Println("# TYPE total_wh gauge")
	fmt.Println("total_wh{" + tags + "} " + strconv.Itoa(totalWh))
}

func queryDevice(connection net.Conn, request string) []byte {
	scrambledText := scramble([]byte(request))
	bytesWritten, writeErr := fmt.Fprint(connection, string(scrambledText))
	if writeErr != nil || bytesWritten != len(scrambledText) {
		panic("Could not write command to connection")
	}
	setDeadlineErr := connection.SetReadDeadline(time.Now().Add(1 * time.Second))
	if setDeadlineErr != nil {
		panic("Could not set read timeout on connection: " + setDeadlineErr.Error())
	}
	buffer := make([]byte, 1024)
	bytesRead, readErr := bufio.NewReader(connection).Read(buffer)
	if readErr != nil {
		panic("Could not read from connection: " + readErr.Error())
	}
	return unscramble(buffer[:MinInt(bytesRead, len(buffer))])
}

func OpenConnection(host string, port uint16) net.Conn {
	connection, connErr := net.Dial("tcp", host+":"+strconv.Itoa(int(port)))
	if connErr != nil {
		panic("Could not open connection: " + connErr.Error())
	}
	return connection
}

func MinInt(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func scramble(b []byte) []byte {
	var iv byte = 171
	buffer := make([]byte, 4+len(b))

	WriteUInt32ToBufferBigEndian(buffer, uint32(len(b)))
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

func WriteUInt32ToBufferBigEndian(b []byte, i uint32) {
	b[0] = byte((i >> 24) & 0xff)
	b[1] = byte((i >> 16) & 0xff)
	b[2] = byte((i >> 8) & 0xff)
	b[3] = byte(i & 0xff)
}
