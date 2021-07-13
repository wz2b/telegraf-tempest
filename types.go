package main

/*****************************************************************************
 * Tempest UDP json objects
 * Built from documentation of the format here:
 *     https://weatherflow.github.io/Tempest/api/udp/v143/
 ****************************************************************************/

import (
	"errors"
	"time"
)

type TempestMessage struct {
	SerialNumber string `json:"serial_number"`
	MessageType  string `json:"type"`
}

type RapidWind struct {
	TempestMessage
	HubSN       string    `json:"hub_sn"`
	Observation []float64 `json:"ob"`
}

func (w *RapidWind) IsValid() bool {
	return len(w.Observation) >= 2
}

func (w *RapidWind) Speed() (float64, error) {
	return w.GetField(1)
}

func (w *RapidWind) Direction() (float64, error) {
	return w.GetField(2)
}

func (w *RapidWind) GetField(n int) (float64, error) {
	if len(w.Observation) < n {
		return 0, errors.New("No such field")
	}
	return w.Observation[n], nil
}

type HubStatus struct {
	TempestMessage
	SerialNumber     string  `json:"serial_number"`
	FirmwareRevision string  `json:"firmware_revision"`
	Uptime           int64   `json:"uptime"`
	RSSI             float64 `json:"rssi"`
	Timestamp        uint64  `json:"timestamp"`
	ResetFlags       string  `json:"reset_flags"`
	Seq              int64   `json:"seq"`
	RadioStats       []int64 `json:"radio_stats"`
	// the 'fs' and 'mqtt_stats' fields are for internal use only and is not decoded.
}

type DeviceStatus struct {
	TempestMessage
	SerialNumber     string  `json:"serial_number"`
	Timestamp        int64   `json:"timestamp"`
	Uptime           int64   `json:"uptime"`
	Voltage          float64 `json:"voltage"`
	FirmwareRevision int32   `json:"firmware_revision"`
	RSSI             float64 `json:"rssi"`
	HubRSSI          float64 `json:"hub_rssi"`
	SensorStatus     int32   `json:"sensor_status"`
}

type LightningStrikeEvent struct {
	TempestMessage
	Evt []float64 `json:"evt"`
}

func (l *LightningStrikeEvent) GetTime() *time.Time {
	if l == nil || len(l.Evt) < 1 {
		return nil
	}

	fv := l.Evt[0]
	u := int64(fv)
	t := time.Unix(u, 0)
	return &t
}

func (l *LightningStrikeEvent) GetDistanceKm() *float64 {
	if l == nil || len(l.Evt) < 2 {
		return nil
	}
	return &l.Evt[1]
}

func (l *LightningStrikeEvent) GetStrikeEnergy() *float64 {
	if l == nil || len(l.Evt) < 3 {
		return nil
	}
	return &l.Evt[2]
}

/*
 * The station observation struct is a header plus an array of arrays of numbers.  Each
 * one of these reports can contain multiple observations (though I have never seen it
 * actually happen).  If thought of as a table, the rows are individual observations
 * and the columns are individual fields like wind speed, temperature, etc.
 *
 * Users of this object can pull the individual rows apart using the helper functions
 * below.
 */
type StationObservation struct {
	TempestMessage
	SerialNumber string      `json:"serial_number"`
	Observations [][]float64 `json:"obs"`
}

func (o *StationObservation) GetField(obs int, n int) (float64, error) {
	if len(o.Observations[obs]) < n {
		return 0, errors.New("No such field")
	}
	return o.Observations[obs][n], nil
}

func (o *StationObservation) NumObservations() int {
	return len(o.Observations)
}

func (o *StationObservation) Time(obs int) (*time.Time, error) {
	fv, err := o.GetField(obs, 0)
	if err != nil {
		return nil, err
	}

	u := int64(fv)
	t := time.Unix(u, 0)

	return &t, nil
}

func (o *StationObservation) WindLull(obs int) (float64, error) {
	return o.GetField(obs, 1)
}

func (o *StationObservation) WindAvg(obs int) (float64, error) {
	return o.GetField(obs, 2)
}

func (o *StationObservation) WindGust(obs int) (float64, error) {
	return o.GetField(obs, 3)
}

func (o *StationObservation) WindDir(obs int) (float64, error) {
	return o.GetField(obs, 4)
}

func (o *StationObservation) WindSampleInterval(obs int) (float64, error) {
	return o.GetField(obs, 5)
}

func (o *StationObservation) AirTemp(obs int) (float64, error) {
	return o.GetField(obs, 7)
}

func (o *StationObservation) StationPressure(obs int) (float64, error) {
	return o.GetField(obs, 6)
}

func (o *StationObservation) RelativeHumidity(obs int) (float64, error) {
	return o.GetField(obs, 8)
}

func (o *StationObservation) Illuminance(obs int) (float64, error) {
	return o.GetField(obs, 9)
}

func (o *StationObservation) UV(obs int) (float64, error) {
	return o.GetField(obs, 10)
}

func (o *StationObservation) SolarRadiation(obs int) (float64, error) {
	return o.GetField(obs, 11)
}

func (o *StationObservation) RainPreviousMinute(obs int) (float64, error) {
	return o.GetField(obs, 12)
}

func (o *StationObservation) PrecipitationType(obs int) (float64, error) {
	return o.GetField(obs, 13)
}

func (o *StationObservation) LightningStrikeAverageDistance(obs int) (float64, error) {
	return o.GetField(obs, 14)
}

func (o *StationObservation) LightningStrikeCount(obs int) (float64, error) {
	return o.GetField(obs, 15)
}

func (o *StationObservation) Battery(obs int) (float64, error) {
	return o.GetField(obs, 16)
}

func (o *StationObservation) ReportInterval(obs int) (float64, error) {
	return o.GetField(obs, 17)
}
