module telegraf-tempest

go 1.15

require (
	github.com/go-kit/kit v0.11.0
	github.com/influxdata/line-protocol v0.0.0-20210311194329-9aa0e372d097
	github.com/wz2b/telegraf-execd-toolkit v0.0.0-unpublished
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/wz2b/telegraf-execd-toolkit v0.0.0-unpublished => ../telegraf-execd-toolkit
