package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func registerNewMetrics(pushGatewayUrl string, tags string, state int, onTime time.Duration, voltageMv float64,
	currentMa float64, powerMw float64, totalWh float64) {

	payload := fmt.Sprintf(`
# TYPE state gauge
state{%[1]s} %[2]d
# TYPE on_time gauge
on_time{%[1]s} %.3[3]f
# TYPE voltage_mv gauge
voltage_mv{%[1]s} %.3[4]f
# TYPE current_ma gauge
current_ma{%[1]s} %.3[5]f
# TYPE power_mw gauge
power_mw{%[1]s} %.3[6]f
# TYPE total_wh gauge
total_wh{%[1]s} %.3[7]f
`, tags, state, onTime.Seconds(), voltageMv, currentMa, powerMw, totalWh)

	// fmt.Println(payload)

	client := &http.Client{
		Timeout: time.Second * 1,
	}
	response, err := client.Post(pushGatewayUrl, "text/plain", bytes.NewBufferString(payload))
	if err != nil {
		panic("Could not post metrics to push gateway: " + err.Error())
	}
	if response.StatusCode != 202 {
		panic("Metrics were not accepted by push gateway: " + response.Status)
	}
}

func deleteMetrics(pushGatewayUrl string) {
	client := &http.Client{
		Timeout: time.Second * 1,
	}
	request, err := http.NewRequest("DELETE", pushGatewayUrl, nil)
	if err != nil {
		fmt.Println("Could not delete metrics from gateway: " + err.Error())
		return
	}

	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Could not delete metrics from gateway: " + err.Error())
		return
	}
	if response.StatusCode != 202 {
		fmt.Println("Could not delete metrics from gateway: " + response.Status)
		return
	}
}
