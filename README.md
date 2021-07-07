# Tempest Weatherflow Plugin for Telegraf

This application connects the [Weatherflow Tempest](https://weatherflow.com/tempest-weather-system/)
weather station with [Telegraf](https://www.influxdata.com/time-series-platform/telegraf/),
the open-source agent that is a companion to the
[InfluxDB](https://www.influxdata.com/products/influxdb/)
time series database.  This plugin listens to the
[Tempest UDP v143](https://weatherflow.github.io/Tempest/api/udp/v143/) format messages
broadcast to the local network to which the weatherflow hub is connected.  It does not
connect back to the Weatherflow cloud API (it is all local).

Telegraf is fairly general-purpose and telegraf can route data other places than
just influxdb.  It can route data to a few other time series databases, and also
to other destinations such as an MQTT endpoint.  There are a lot of 
[Telegraf Plugins](https://docs.influxdata.com/telegraf/v1.19/plugins/)
available that do different things.

This 'plugin' is an external program that writes data in the
[Influx Line Protocol](https://docs.influxdata.com/influxdb/cloud/reference/syntax/line-protocol/)
format suitable for writing to telegraph using the
[inputs.execd](https://github.com/influxdata/telegraf/blob/release-1.19/plugins/inputs/execd/README.md)
 plugin.  It takes no parameters, and starts listening on the local network for UDP broadcasts
 to port 50222.  To use it, create a telegraf input configuration like this:
 
 ```yaml
[[inputs.execd]]
  command = [ "tempest_udp" ]
```

That's pretty much all you have to do, but if that's all you have the data won't go anywhere.
A more complete example is:

```yaml
[[aggregators.merge]]
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = true

[[outputs.influxdb_v2]]
  urls = ["http://127.0.0.1:8086"]
  bucket = "power"
  organization = "your_org"
  token = "your_influx_token"

[[inputs.execd]]
  ## Commands array
  command = [ "tempest_udp" ]

```

Currently this writes the following measurements:

* wind - rapid wind observation
* observation - general weather information (temperature, pressure, humidity, etc)
* hub_status - periodic messages coming from the hub telling you it is alive.  There is an RSSI
in this message and they don't document what it means, but it may be the RSSI the Hub is reporting for 
whatever Wi-Fi network it is conneted to
* device_status - info about the remote device, including two RSSIs which are also undocumented,
but it's likely that one of them is tempest-to-hub and the other is hub-to-tempest
 
#### Error Output
Logging in go is a mess.

This project includes a lightweight log adapter that attempts to put output log messages
with a leading "comment" indicator so that telegraf will ignore them.  This works if you
set the logger to write to os.Stdout, but if you do that you get no feedback at all.
On the other hand, if you configure it to write to os.Stderr telegraf sees it as an actual
error, even though some of what's output may be status or debug messages. 

#### Future Improvements
 * Make the individual measurement names be configurable
 * Make logging a little more configurable, maybe through a -debug or -quiet flag
 * Add lightning strike observation (have yet to see it happen, though, so it's not there yet)
 * If you need/want something else, write an issue and I'll see what I can do

# Author
    Christopher Piggott