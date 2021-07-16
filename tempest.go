package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"telegraf-tempest/internal/tclogger"
	"time"
)

var telegrafLogger = tclogger.Create()

func main() {

	/*
	 * Install the telegraf compatible log writer
	 */
	telegrafLogger.Writer = os.Stdout
	telegrafLogger.Start()

	sock, err := net.ListenPacket("udp", ":50222")
	if err != nil {
		log.Fatal(err)
	}

	byteBuf := make([]byte, 2000)
	for {
		n, _, err := sock.ReadFrom(byteBuf)
		if err != nil {
			log.Fatal("Could not read from socket", err)
		}
		bytes := byteBuf[:n]

		/*
		 * Decode the envelope.  This will tell us what type of message this is.
		 */
		var message TempestMessage
		err = json.Unmarshal(bytes, &message)
		if err != nil {
			log.Fatal("Error unmarshalling message", err)
		}

		if message.MessageType == "rapid_wind" {
			writeRapidWind(os.Stdout, bytes)
		} else if message.MessageType == "hub_status" {
			writeHubStatus(os.Stdout, bytes)
		} else if message.MessageType == "device_status" {
			writeDeviceStatus(os.Stdout, bytes)
		} else if message.MessageType == "obs_st" {
			writeStationObservation(os.Stdout, bytes)
		} else if message.MessageType == "evt_strike" {
			writeLightningStrike(os.Stdout, bytes)
		} else {
			log.Printf("Received: unknown message %s\n", message.MessageType)
			log.Print(string(byteBuf[:n]))
		}

	}

}

func writeLightningStrike(ostream *os.File, bytes []byte) {
	var strike LightningStrikeEvent
	err := json.Unmarshal(bytes, &strike)
	if err != nil {
		log.Print(err)
		log.Print(string(bytes))
	} else {
		metric, _ := CreateTempestMetricNow("lightning_strike")
		metric.AddTag("station", strike.SerialNumber)

		metric.AddField("distance_km", strike.GetDistanceKm())
		metric.AddField("energy", strike.GetStrikeEnergy())
		_, err := metric.WriteTo(ostream)
		if err != nil {
			log.Print(err)
		}
	}
}

func writeHubStatus(ostream io.Writer, bytes []byte) {
	var hub HubStatus
	err := json.Unmarshal(bytes, &hub)
	if err != nil {
		log.Print(err)
		log.Print(string(bytes))
	} else {
		metric, _ := CreateTempestMetricNow("hub_status")
		metric.AddTag("hub", hub.SerialNumber)
		metric.AddField("seq", hub.Seq)
		metric.AddField("rssi", hub.RSSI)
		_, err := metric.WriteTo(ostream)
		if err != nil {
			log.Print(err)
		}
	}
}

func writeRapidWind(ostream io.Writer, ByteBuf []byte) {
	var wind RapidWind
	err := json.Unmarshal(ByteBuf, &wind)
	if err != nil {
		log.Print(err)
		log.Print(string(ByteBuf))
	} else {

		metric, _ := CreateTempestMetricNow("wind")
		v, e := wind.Speed()

		metric.AddTag("hub", wind.HubSN)
		metric.AddTag("station", wind.SerialNumber)

		metric.AddFieldIfValid("wind_speed", v, e)

		v, e = wind.Direction()
		metric.AddFieldIfValid("wind_dir", v, e)
		_, err := metric.WriteTo(ostream)
		if err != nil {
			log.Print(err)
		}
	}
}

func writeStationObservation(ostream io.Writer, bytes []byte) {
	var observation StationObservation
	err := json.Unmarshal(bytes, &observation)
	if err != nil {
		log.Print(err)
		log.Print(string(bytes))
		return
	}
	now := time.Now()
	numObs := observation.NumObservations()
	for obs := 0; obs < numObs; obs++ {

		t, err := observation.Time(obs)

		if err != nil {
			log.Println("Unable to get observation time")
			return
		}

		if t == nil {
			t = &now
		}
		metric, _ := CreateTempestMetric("observation", *t)
		metric.AddTag("station", observation.SerialNumber)

		v, e := observation.AirTemp(obs)
		metric.AddFieldIfValid("temperature", v, e)

		v, e = observation.RelativeHumidity(obs)
		metric.AddFieldIfValid("humidity", v, e)

		v, e = observation.StationPressure(obs)
		metric.AddFieldIfValid("pressure", v, e)

		v, e = observation.WindAvg(obs)
		metric.AddFieldIfValid("wind_spd", v, e)

		v, e = observation.WindGust(obs)
		metric.AddFieldIfValid("wind_gust", v, e)

		v, e = observation.WindLull(obs)
		metric.AddFieldIfValid("wind_lull", v, e)

		v, e = observation.WindDir(obs)
		metric.AddFieldIfValid("wind_dir", v, e)

		v, e = observation.RainPreviousMinute(obs)
		metric.AddFieldIfValid("rain_previous_min", v, e)

		v, e = observation.LightningStrikeCount(obs)
		metric.AddFieldIfValid("lightning_strikes", v, e)

		v, e = observation.UV(obs)
		metric.AddFieldIfValid("uv", v, e)

		v, e = observation.Illuminance(obs)
		metric.AddFieldIfValid("illuminance", v, e)

		v, e = observation.SolarRadiation(obs)
		metric.AddFieldIfValid("solar_radiation", v, e)

		_, err = metric.WriteTo(ostream)
		if err != nil {
			log.Print(err)
		}
	}
}

func writeDeviceStatus(ostream io.Writer, bytes []byte) {
	var status DeviceStatus
	err := json.Unmarshal(bytes, &status)
	if err != nil {
		log.Print(err)
		log.Print(string(bytes))
		return
	}
	metric, _ := CreateTempestMetricNow("device_status")
	metric.AddTag("station", status.SerialNumber)
	metric.AddField("sensor_status", status.SensorStatus)
	metric.AddField("rssi", status.RSSI)
	metric.AddField("hub_rssi", status.HubRSSI)
	metric.AddField("battery", status.Voltage)
	_, err = metric.WriteTo(ostream)
	if err != nil {
		log.Print(err)
	}
}
