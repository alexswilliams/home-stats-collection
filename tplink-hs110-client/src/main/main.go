package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	pollTime, host, port, pushGatewayUrl := getEnvVars()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	ticker := time.NewTicker(pollTime)
	pleaseExit := make(chan bool, 1)
	exited := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-pleaseExit:
				exited <- true
				return
			case <-ticker.C:
				updateMetrics(pushGatewayUrl, host, port)
			}
		}
	}()

	sig := <-signals
	fmt.Println("Received signal: " + sig.String())

	pleaseExit <- true
	fmt.Println("Waiting for current request to finish...")
	<-exited
	fmt.Println("Exiting")
	os.Exit(0)

}

var commands = map[string]map[string]string{
	"system": {
		"get_sysinfo": `{"system":{"get_sysinfo":null}}`,
	},
	"emeter": {
		"get_realtime":    `{"emeter":{"get_realtime":{}}}`,
		"get_vgain_igain": `{"emeter":{"get_vgain_igain":{}}}`,
	},
}

func updateMetrics(pushGatewayUrl string, host string, port uint16) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(time.Now().String() + " - Failed to query device: " + err.(string))
			deleteMetrics(pushGatewayUrl)
		}
	}()
	connection := openConnection(host, port)
	defer func() { _ = connection.Close() }()

	deviceInfo := queryDevice(connection, commands["system"]["get_sysinfo"])
	alias, id, mac, state, onTime := extractDeviceInfo(deviceInfo)

	realTimeInfo := queryDevice(connection, commands["emeter"]["get_realtime"])
	voltageMv, currentMa, powerMw, totalWh := extractRealTimeInfo(realTimeInfo)

	tags := `alias="` + alias + `",id="` + id + `",mac="` + mac + `"`
	registerNewMetrics(pushGatewayUrl, tags, state, onTime, voltageMv, currentMa, powerMw, totalWh)
}

func extractRealTimeInfo(realTimeInfo []byte) (voltageMv float64, currentMa float64, powerMw float64, totalWh float64) {
	var realtimeJson map[string]map[string]map[string]float64
	if err := json.Unmarshal(realTimeInfo, &realtimeJson); err != nil {
		panic("Could not unmarshal real-time JSON string: " + err.Error())
	}
	voltageMv = realtimeJson["emeter"]["get_realtime"]["voltage_mv"]
	currentMa = realtimeJson["emeter"]["get_realtime"]["current_ma"]
	powerMw = realtimeJson["emeter"]["get_realtime"]["power_mw"]
	totalWh = realtimeJson["emeter"]["get_realtime"]["total_wh"]
	return
}

func extractDeviceInfo(deviceInfo []byte) (alias string, id string, mac string, state int, onTime time.Duration) {
	var infoJson map[string]map[string]map[string]interface{}
	if err := json.Unmarshal(deviceInfo, &infoJson); err != nil {
		panic("Could not unmarshal info JSON string: " + err.Error())
	}
	alias = infoJson["system"]["get_sysinfo"]["alias"].(string)
	id = infoJson["system"]["get_sysinfo"]["deviceId"].(string)
	mac = infoJson["system"]["get_sysinfo"]["mac"].(string)
	state = int(infoJson["system"]["get_sysinfo"]["relay_state"].(float64))
	onTime = time.Duration(int(infoJson["system"]["get_sysinfo"]["on_time"].(float64))) * time.Second
	return
}
