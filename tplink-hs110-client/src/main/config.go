package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func getEnvVars() (pollTime time.Duration, host string, port uint16, pushGatewayUrl string) {
	pollTime = getPollTime()
	host = getHost()
	port = getPort()
	pushGatewayUrl = getPushGatewayUrl()
	fmt.Printf(`Polling TPLink HS110 with the following configuration:
 • Poll Frequency: %v
 • Host: %s
 • Port: %d
 • Prometheus Push Gateway: %s
`,
		pollTime, host, port, pushGatewayUrl)
	return
}

func getPollTime() time.Duration {
	pollTimeString, found := os.LookupEnv("POLL_TIME_SECONDS")
	if found {
		pollTimeParsed, err := strconv.ParseUint(pollTimeString, 10, 16)
		if err != nil {
			panic("Could not parse POLL_TIME_SECONDS: " + err.Error())
		}
		return time.Second * time.Duration(pollTimeParsed)
	} else {
		return time.Second * 5
	}
}

func getHost() string {
	host, found := os.LookupEnv("TPLINK_HOST")
	if found {
		if host == "" {
			panic("TPLINK_HOST cannot be empty if set.")
		}
		return host
	} else {
		panic("Env var TPLINK_HOST must be set.")
	}
}

func getPort() uint16 {
	portString, found := os.LookupEnv("TPLINK_PORT")
	if found {
		portParsed, err := strconv.ParseUint(portString, 10, 16)
		if err != nil {
			panic("Could not parse POLL_TIME_SECONDS: " + err.Error())
		}
		return uint16(portParsed)
	} else {
		return 9999
	}
}

func getPushGatewayUrl() string {
	url, found := os.LookupEnv("PUSH_GW_URL")
	if found {
		if url == "" {
			panic("PUSH_GW_URL cannot be empty if set.")
		}
		return url
	} else {
		panic("Env var PUSH_GW_URL must be set.")
	}
}
