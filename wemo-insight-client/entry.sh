#!/usr/bin/env sh

PUSH_ADDRESS="http://prom-push-gw:9091/metrics/job/wemo-scraper"
TEMP_FILE="/tmp/out.txt"
ADDRESS=${WEMO_ADDRESS:-192.168.1.25}
PORT=${WEMO_PORT:-49153}
POLL_TIME=${WEMO_POLL_TIME:-5}

echo "Scraping from ${ADDRESS}:${PORT} every ${POLL_TIME} second(s)"
echo "Saving to ${TEMP_FILE}"
echo "Pushing to ${PUSH_ADDRESS}"

function deleteMetrics {
    curl -s -X DELETE ${PUSH_ADDRESS}
    local success=$?
    if [[ ${success} -ne 0 ]]; then
        echo "Could not delete metrics from push server"
    fi
}

function scrapeWeMo {
    python main.py ${ADDRESS} ${PORT} > ${TEMP_FILE}
    local success=$?
    if [[ ${success} -ne 0 ]]; then
        echo "Could not scrape WeMo - deleting old metrics from push server."
        deleteMetrics
    fi
}

function postMetrics {
    curl -s --data-binary @${TEMP_FILE} ${PUSH_ADDRESS}
    local success=$?
    if [[ ${success} -ne 0 ]]; then
        echo "Could not push metrics to push server"
    fi
}


while true; do
    scrapeWeMo
    postMetrics
    # echo "Pushed - sleeping ${POLL_TIME} seconds"

    sleep ${POLL_TIME}
done