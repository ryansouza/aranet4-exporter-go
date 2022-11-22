package aranet

import (
	"bytes"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"sbinet.org/x/aranet4"
)

type mock struct {
	data aranet4.Data
	room string
}

func (m mock) Read() aranet4.Data {
	return m.data
}

func (m mock) Room() string {
	return m.room
}

func TestCollector(t *testing.T) {
	aTime := time.Unix(1655701702, 0)

	mockAranet1 := mock{room: "one", data: aranet4.Data{
		H:        60.0,
		P:        1001.0,
		T:        20.2,
		CO2:      456,
		Battery:  90.0,
		Quality:  1,
		Interval: 2 * time.Minute,
		Time:     aTime,
	}}

	mockAranet2 := mock{room: "two", data: aranet4.Data{
		H:        90.0,
		P:        1801.0,
		T:        42.42,
		CO2:      2020,
		Battery:  13.0,
		Quality:  3,
		Interval: 5 * time.Minute,
		Time:     aTime,
	}}

	collector := Collector{ReportOldData: true, Aranets: []AranetData{mockAranet1, mockAranet2}}

	registry := prometheus.NewRegistry()
	registry.MustRegister(&collector)

	gatheredMetrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	out := &bytes.Buffer{}
	for _, mf := range gatheredMetrics {
		if _, err := expfmt.MetricFamilyToText(out, mf); err != nil {
			t.Fatalf("metrics to text failed: %v", err)
		}
	}

	expected := `# HELP aranet4_battery_percent Battery level as percentage.
# TYPE aranet4_battery_percent gauge
aranet4_battery_percent{room="one"} 90
aranet4_battery_percent{room="two"} 13
# HELP aranet4_co2_ppm CO2 level in parts per million.
# TYPE aranet4_co2_ppm gauge
aranet4_co2_ppm{room="one"} 456
aranet4_co2_ppm{room="two"} 2020
# HELP aranet4_collector_last_scrape Last scrape time as unix time.
# TYPE aranet4_collector_last_scrape gauge
aranet4_collector_last_scrape{room="one"} 1.655701702e+09
aranet4_collector_last_scrape{room="two"} 1.655701702e+09
# HELP aranet4_collector_next_scrape Last scrape time as unix time.
# TYPE aranet4_collector_next_scrape gauge
aranet4_collector_next_scrape{room="one"} 1.655701822e+09
aranet4_collector_next_scrape{room="two"} 1.655702002e+09
# HELP aranet4_humidity_percent Relative humidity as percentage.
# TYPE aranet4_humidity_percent gauge
aranet4_humidity_percent{room="one"} 60
aranet4_humidity_percent{room="two"} 90
# HELP aranet4_pressure_hpa Atmospheric pressure in hPa.
# TYPE aranet4_pressure_hpa gauge
aranet4_pressure_hpa{room="one"} 1001
aranet4_pressure_hpa{room="two"} 1801
# HELP aranet4_temperature_c Temperature in celcius.
# TYPE aranet4_temperature_c gauge
aranet4_temperature_c{room="one"} 20.2
aranet4_temperature_c{room="two"} 42.42
`
	got := out.String()

	if got != expected {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expected, got, false)

		t.Fatalf("collector output did not match:\n%s", dmp.DiffPrettyText(diffs))
	}
}
