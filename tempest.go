package main

import (
	"encoding/json"
	"fmt"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	line_metric_encoder "github.com/wz2b/telegraf-execd-toolkit/line-metric-encoder"
	tlogger "github.com/wz2b/telegraf-execd-toolkit/telegraf-logger"
	"io"
	"net"
	"os"
	"time"
)

var config TempestAgentConfig

const DAYS = 24 * time.Hour

var klog kitlog.Logger

func main() {
	logFactory, err := tlogger.NewTelegrafLoggerConfiguration(true)

	if err != nil {
		panic(err)
	}

	klog := logFactory.Create()

	mp := line_metric_encoder.NewMetricEncoderPool()

	sock, err := net.ListenPacket("udp", ":50222")
	if err != nil {
		level.Error(klog).Log("msg", "Unable to listen()", "error", err)
		os.Exit(1)
	}

	byteBuf := make([]byte, 2000)
	for {
		n, _, err := sock.ReadFrom(byteBuf)
		if err != nil {
			level.Error(klog).Log("msg", "Could not read from socket", "error", err)
			os.Exit(1)
		}
		bytes := byteBuf[:n]

		/*
		 * Decode the envelope.  This will tell us what type of message this is.
		 */
		var message TempestMessage
		err = json.Unmarshal(bytes, &message)
		if err != nil {
			level.Warn(klog).Log("msg", "Error unmarshalling message", "error", err)
			continue
		}

		if message.MessageType == "rapid_wind" {
			writeRapidWind(mp, os.Stdout, bytes)
		} else if message.MessageType == "hub_status" {
			writeHubStatus(mp, os.Stdout, bytes)
		} else if message.MessageType == "device_status" {
			writeDeviceStatus(mp, os.Stdout, bytes)
		} else if message.MessageType == "obs_st" {
			writeStationObservation(mp, os.Stdout, bytes)
		} else if message.MessageType == "evt_strike" {
			writeLightningStrike(mp, os.Stdout, bytes)
		} else {
			klog.Log("msg",
				fmt.Sprintf("Received: unknown message type %s\n", message.MessageType))
		}
	}
}

func writeLightningStrike(mp *line_metric_encoder.MetricEncoderPool, ostream *os.File, bytes []byte) {
	var strike LightningStrikeEvent
	err := json.Unmarshal(bytes, &strike)
	if err != nil {
		level.Debug(klog).Log("msg", "Unable to parse evt_strike", "error", err)
	} else {
		metric := mp.NewMetric("lightning_strike")
		metric.AddTag("station", strike.SerialNumber)

		metric.AddField("distance_km", strike.GetDistanceKm())
		metric.AddField("energy", strike.GetStrikeEnergy())
		_, err := metric.Write(ostream)
		if err != nil {
			level.Warn(klog).Log("msg", "unable to write evt_strike metric", "error", err)
		}
	}
}

func writeHubStatus(mp *line_metric_encoder.MetricEncoderPool, ostream io.Writer, bytes []byte) {
	var hub HubStatus
	err := json.Unmarshal(bytes, &hub)
	if err != nil {
		level.Warn(klog).Log("msg", "unable to parse hub status event", "error", err)
	} else {
		metric := mp.NewMetric("hub_status")
		metric.WithTag("hub", hub.SerialNumber).WithField("seq", hub.Seq).WithField("rssi", hub.RSSI)

		_, err := metric.Write(ostream)
		if err != nil {
			level.Warn(klog).Log("msg", "unable to write hub status metric", "error", err)
		}
	}
}

func writeRapidWind(mp *line_metric_encoder.MetricEncoderPool, ostream io.Writer, ByteBuf []byte) {
	var wind RapidWind
	err := json.Unmarshal(ByteBuf, &wind)
	if err != nil {
		level.Warn(klog).Log("msg", "unable to parse rapid wind event", "error", err)
	} else {

		metric := mp.NewMetric("wind")
		v, e := wind.Speed()

		metric.WithTag("hub", wind.HubSN)
		metric.WithTag("station", wind.SerialNumber)

		addFieldIfNoError(metric, "wind_speed", v, e)

		v, e = wind.Direction()
		addFieldIfNoError(metric, "wind_dir", v, e)
		_, err := metric.Write(ostream)
		if err != nil {
			level.Warn(klog).Log("msg", "unable to write hub status metric", "error", err)
		}
	}
}

func addFieldIfNoError(metric *line_metric_encoder.WrappedMetric, key string, value interface{}, err error) {
	if err != nil {
		level.Warn(klog).Log("msg", "unable to add field", "error", err)
	} else {
		metric.WithField(key, value)
	}
}

func writeStationObservation(mp *line_metric_encoder.MetricEncoderPool, ostream io.Writer, bytes []byte) {
	var observation StationObservation
	err := json.Unmarshal(bytes, &observation)
	if err != nil {
		level.Warn(klog).Log("msg", "unable to parse station observation event", "error", err)
		return
	}
	now := time.Now()
	numObs := observation.NumObservations()
	for obs := 0; obs < numObs; obs++ {

		t, err := observation.Time(obs)

		if err != nil {
			level.Warn(klog).Log("msg", "unable to parse timestamp from observation", "error", err)
			return
		}

		if t == nil {
			t = &now
		}
		metric := mp.NewMetric("observation").WithTime(*t)
		metric.AddTag("station", observation.SerialNumber)

		v, e := observation.AirTemp(obs)
		addFieldIfNoError(metric, "temperature", v, e)

		v, e = observation.RelativeHumidity(obs)
		addFieldIfNoError(metric, "humidity", v, e)

		v, e = observation.StationPressure(obs)
		addFieldIfNoError(metric, "pressure", v, e)

		v, e = observation.WindAvg(obs)
		addFieldIfNoError(metric, "wind_spd", v, e)

		v, e = observation.WindGust(obs)
		addFieldIfNoError(metric, "wind_gust", v, e)

		v, e = observation.WindLull(obs)
		addFieldIfNoError(metric, "wind_lull", v, e)

		v, e = observation.WindDir(obs)
		addFieldIfNoError(metric, "wind_dir", v, e)

		v, e = observation.RainPreviousMinute(obs)
		addFieldIfNoError(metric, "rain_previous_min", v, e)

		v, e = observation.LightningStrikeCount(obs)
		addFieldIfNoError(metric, "lightning_strikes", v, e)

		v, e = observation.UV(obs)
		addFieldIfNoError(metric, "uv", v, e)

		v, e = observation.Illuminance(obs)
		addFieldIfNoError(metric, "illuminance", v, e)

		v, e = observation.SolarRadiation(obs)
		addFieldIfNoError(metric, "solar_radiation", v, e)

		_, err = metric.Write(ostream)
		if err != nil {
			level.Warn(klog).Log("msg", "unable to write station observation metric", "error", err)

		}
	}
}

func writeDeviceStatus(mp *line_metric_encoder.MetricEncoderPool, ostream io.Writer, bytes []byte) {
	var status DeviceStatus
	err := json.Unmarshal(bytes, &status)
	if err != nil {
		level.Warn(klog).Log("msg", "unable to parse device status event", "error", err)

		return
	}
	metric := mp.NewMetric("device_status")
	metric.AddTag("station", status.SerialNumber)
	metric.AddField("sensor_status", status.SensorStatus)
	metric.AddField("rssi", status.RSSI)
	metric.AddField("hub_rssi", status.HubRSSI)
	metric.AddField("battery", status.Voltage)
	_, err = metric.Write(ostream)
	if err != nil {
		level.Warn(klog).Log("msg", "unable to write device status metric", "error", err)
	}
}
