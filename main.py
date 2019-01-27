import requests
import re
import time
import sys

if len(sys.argv) >= 2:
    address = sys.argv[1]
else:
    address = "192.168.1.25"

if len(sys.argv) >= 3:
    port = sys.argv[2]
else:
    port = 49153


def get_insight_params(addr, prt):
    url = "http://" + addr + ":" + str(prt) + "/upnp/control/insight1"

    payload = """<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    <u:GetInsightParams xmlns:u="urn:Belkin:service:insight:1"></u:GetInsightParams>
  </s:Body>
</s:Envelope>"""

    headers = {
        'Content-Type': 'text/xml; charset="utf-8"',
        'SOAPACTION': '"urn:Belkin:service:insight:1#GetInsightParams"'
    }

    response = requests.request("POST", url, data=payload, headers=headers)

    if not response.ok:
        print("Response was not 200 OK: " + response.status_code)
        print(response.text)
        exit(1)

    regex = re.compile('.*<InsightParams>(.*)</InsightParams>.*', re.MULTILINE)
    # <InsightParams>8|1548203519|100|1601|415096|1209600|668|615|35238849|6359688716.000000|30000</InsightParams>

    line = regex.match(response.text.replace("\n", "").replace("\r", ""))
    if line is None:
        print("Response did not contain expected information")
        print(response.text)
        exit(1)

    array = line.group(1).split("|")

    def state(x):
        if x == 0:
            return "Off"
        elif x == 1:
            return "On"
        elif x == 8:
            return "Standby"
        else:
            return "Unknown"

    return {
        'requestTimeEpochSeconds': int(time.time()),
        'state': state(int(array[0])),
        'stateVal': int(array[0]),
        'lastChangeEpochSeconds': int(array[1]),
        'onForSeconds': int(array[2]),
        'onForTodaySeconds': int(array[3]),
        'onOverallSeconds': int(array[4]),
        'totalOverallSeconds': int(array[5]),
        'mysteryNumber': int(array[6]),
        'currentPowerMilliWatts': int(array[7]),
        'powerTodayMilliWatts': int(array[8]),
        'overallPowerMilliWatts': int(float(array[9])),
        'standbyThresholdMilliWatts': int(array[10])
    }


def get_extra_meta_data(addr, prt):
    url = "http://" + addr + ":" + str(prt) + "/upnp/control/metainfo1"

    payload = """<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    <u:GetExtMetaInfo xmlns:u="urn:Belkin:service:metainfo:1"></u:GetExtMetaInfo>
  </s:Body>
</s:Envelope>"""

    headers = {
        'Content-Type': 'text/xml; charset="utf-8"',
        'SOAPACTION': '"urn:Belkin:service:metainfo:1#GetExtMetaInfo"'
    }

    response = requests.request("POST", url, data=payload, headers=headers)

    if not response.ok:
        print("Response was not 200 OK: " + response.status_code)
        print(response.text)
        exit(1)

    regex = re.compile('.*<ExtMetaInfo>(.*)</ExtMetaInfo>.*', re.MULTILINE)
    # <ExtMetaInfo>1|0|1|0|1419:43:35|4|1548206463|11600781|1|Insight|4|41|3|30000|1|4</ExtMetaInfo>

    line = regex.match(response.text.replace("\n", "").replace("\r", ""))
    if line is None:
        print("Response did not contain expected information")
        print(response.text)
        exit(1)

    array = line.group(1).split("|")

    return {
        'requestTimeEpochSeconds': int(time.time()),
        'totalOnTime': array[4],
        'deviceTimeEpochSeconds': int(array[6])
    }


insight = get_insight_params(address, port)
# metadata = get_extra_meta_data(address, port)

# log_object = {**insight, **metadata}

log_object = insight

print("# TYPE state gauge")
print("state " + str(log_object["stateVal"]))
print("# TYPE last_change_epoch_seconds counter")
print("last_change_epoch_seconds " + str(log_object["lastChangeEpochSeconds"]))
print("# TYPE on_for_seconds gauge")
print("on_for_seconds " + str(log_object["onForSeconds"]))
print("# TYPE on_for_today_seconds counter")
print("on_for_today_seconds " + str(log_object["onForTodaySeconds"]))
print("# TYPE on_overall_seconds counter")
print("on_overall_seconds " + str(log_object["onOverallSeconds"]))
print("# TYPE mystery_number gauge")
print("mystery_number " + str(log_object["mysteryNumber"]))
print("# TYPE current_power_milli_watts gauge")
print("current_power_milli_watts " + str(log_object["currentPowerMilliWatts"]))
print("# TYPE power_today_milli_watts counter")
print("power_today_milli_watts " + str(log_object["powerTodayMilliWatts"]))
print("# TYPE overall_power_milli_watts counter")
print("overall_power_milli_watts " + str(log_object["overallPowerMilliWatts"]))
print("# TYPE standby_threshold_milli_watts gauge")
print("standby_threshold_milli_watts " + str(log_object["standbyThresholdMilliWatts"]))
print()
