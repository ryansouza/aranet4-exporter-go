package aranet

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	labels         = []string{"room"}
	co2Desc        = prometheus.NewDesc("aranet4_co2_ppm", "CO2 level in parts per million.", labels, nil)
	tempDesc       = prometheus.NewDesc("aranet4_temperature_c", "Temperature in celcius.", labels, nil)
	humidityDesc   = prometheus.NewDesc("aranet4_humidity_percent", "Relative humidity as percentage.", labels, nil)
	pressureDesc   = prometheus.NewDesc("aranet4_pressure_hpa", "Atmospheric pressure in hPa.", labels, nil)
	batteryDesc    = prometheus.NewDesc("aranet4_battery_percent", "Battery level as percentage.", labels, nil)
	lastScrapeDesc = prometheus.NewDesc("aranet4_collector_last_scrape", "Last scrape time as unix time.", labels, nil)
	nextScrapeDesc = prometheus.NewDesc("aranet4_collector_next_scrape", "Last scrape time as unix time.", labels, nil)
)

type Collector struct {
	Aranets       []AranetData
	ReportOldData bool
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- co2Desc
	ch <- tempDesc
	ch <- humidityDesc
	ch <- pressureDesc
	ch <- batteryDesc
	ch <- lastScrapeDesc
	ch <- nextScrapeDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	for _, aranet := range c.Aranets {
		data := aranet.Read()
		room := aranet.Room()

		if c.ReportOldData || time.Since(data.Time) < time.Duration(1.1*float64(data.Interval)) {
			ch <- prometheus.MustNewConstMetric(
				co2Desc,
				prometheus.GaugeValue,
				float64(data.CO2),
				room,
			)
			ch <- prometheus.MustNewConstMetric(
				tempDesc,
				prometheus.GaugeValue,
				data.T,
				room,
			)
			ch <- prometheus.MustNewConstMetric(
				humidityDesc,
				prometheus.GaugeValue,
				data.H,
				room,
			)
			ch <- prometheus.MustNewConstMetric(
				pressureDesc,
				prometheus.GaugeValue,
				data.P,
				room,
			)
			ch <- prometheus.MustNewConstMetric(
				batteryDesc,
				prometheus.GaugeValue,
				float64(data.Battery),
				room,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			lastScrapeDesc,
			prometheus.GaugeValue,
			float64(data.Time.Unix()),
			room,
		)

		ch <- prometheus.MustNewConstMetric(
			nextScrapeDesc,
			prometheus.GaugeValue,
			float64(data.Time.Add(data.Interval).Unix()),
			room,
		)
	}
}
