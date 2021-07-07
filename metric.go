package main

import (
	"bytes"
	lineProtocol "github.com/influxdata/line-protocol"
	"io"
	"time"
)

type TempestMetric struct {
	lineProtocol.MutableMetric
}

func CreateTempestMetric(name string, time time.Time) (*TempestMetric, error) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	metric, err := lineProtocol.New(name, tags, fields, time)
	if err != nil {
		return nil, err
	}

	return &TempestMetric{
		MutableMetric: metric,
	}, nil

}

func CreateTempestMetricNow(name string) (*TempestMetric, error) {
	return CreateTempestMetric(name, time.Now())
}

func (m *TempestMetric) AddFieldIfValid(key string, value float64, err error) {
	if err == nil {
		m.AddField(key, value)
	}
}

func (m *TempestMetric) WriteTo(out io.Writer) (int, error) {
	buf := &bytes.Buffer{}
	serializer := lineProtocol.NewEncoder(buf)
	serializer.SetMaxLineBytes(-1)
	serializer.SetFieldTypeSupport(lineProtocol.UintSupport)

	_, err := serializer.Encode(m)
	if err != nil {
		return 0, err
	}

	return out.Write(buf.Bytes())
}
