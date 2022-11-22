package aranet

import (
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/brutella/hap/service"
	"log"
	"sbinet.org/x/aranet4"
)

var qualityMap = map[int]int{1: 1, 2: 3, 3: 5}

type Accessory struct {
	*accessory.A
	TempSensor      *service.TemperatureSensor
	AranetCO2Sensor *AranetCO2Sensor
	HumiditySensor  *service.HumiditySensor
	BatteryService  *service.BatteryService
}

// NewAranetAccessory returns a Thermometer which implements model.Thermometer.
func NewAranetAccessory(info accessory.Info) *Accessory {
	a := Accessory{}
	a.A = accessory.New(info, accessory.TypeSensor)

	a.TempSensor = service.NewTemperatureSensor()
	a.AddS(a.TempSensor.S)

	a.AranetCO2Sensor = NewAranetCO2Sensor()
	a.AddS(a.AranetCO2Sensor.S)

	a.HumiditySensor = service.NewHumiditySensor()
	a.AddS(a.HumiditySensor.S)

	a.BatteryService = service.NewBatteryService()
	a.AddS(a.BatteryService.S)

	return &a
}

func (acc *Accessory) Update(data aranet4.Data) {
	acc.TempSensor.CurrentTemperature.SetValue(data.T)
	acc.AranetCO2Sensor.CarbonDioxideLevel.SetValue(float64(data.CO2))
	if err := acc.AranetCO2Sensor.AirQuality.SetValue(qualityMap[int(data.Quality)]); err != nil {
		log.Printf("Failed to accept air quality: %v", err)
	}
	acc.HumiditySensor.CurrentRelativeHumidity.SetValue(data.H)
	if err := acc.BatteryService.BatteryLevel.SetValue(data.Battery); err != nil {
		log.Printf("Failed to accept battery level: %v", err)
	}
}

type AranetCO2Sensor struct {
	*service.S

	AirQuality         *characteristic.AirQuality
	CarbonDioxideLevel *characteristic.CarbonDioxideLevel
}

func NewAranetCO2Sensor() *AranetCO2Sensor {
	s := AranetCO2Sensor{}
	s.S = service.New(service.TypeAirQualitySensor)

	s.AirQuality = characteristic.NewAirQuality()
	s.AddC(s.AirQuality.C)

	s.CarbonDioxideLevel = characteristic.NewCarbonDioxideLevel()
	s.AddC(s.CarbonDioxideLevel.C)

	return &s
}
