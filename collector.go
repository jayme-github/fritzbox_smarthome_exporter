package main

import (
	"log"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var genericLabels = []string{"device_id", "device_type", "device_name"}

type fritzCollector struct {
	InfoDesc              *prometheus.Desc
	PresentDesc           *prometheus.Desc
	TemperatureDesc       *prometheus.Desc
	TemperatureOffsetDesc *prometheus.Desc
	EnergyWhDesc          *prometheus.Desc
	PowerWDesc            *prometheus.Desc
	SwitchState           *prometheus.Desc
	SwitchMode            *prometheus.Desc
	SwitchBoxLock         *prometheus.Desc
	SwitchDeviceLock      *prometheus.Desc
}

func (fc *fritzCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- fc.InfoDesc
	ch <- fc.PresentDesc
	ch <- fc.TemperatureDesc
	ch <- fc.TemperatureOffsetDesc
	ch <- fc.EnergyWhDesc
	ch <- fc.PowerWDesc
	ch <- fc.SwitchState
	ch <- fc.SwitchMode
	ch <- fc.SwitchBoxLock
	ch <- fc.SwitchDeviceLock
}

func (fc *fritzCollector) Collect(ch chan<- prometheus.Metric) {
	var err error

	fritzClient.Lock()
	l, err := fritzClient.List()
	fritzClient.Unlock()

	if err != nil {
		log.Println("Unable to collect data:", err)
		return
	}

	for _, dev := range l.Devices {
		ch <- prometheus.MustNewConstMetric(
			fc.InfoDesc,
			prometheus.GaugeValue,
			1.0,
			dev.Identifier,
			dev.Productname,
			dev.Name,
			dev.ID,
			dev.Fwversion,
			dev.Manufacturer,
			dev.Functionbitmask,
		)

		ch <- prometheus.MustNewConstMetric(
			fc.PresentDesc,
			prometheus.GaugeValue,
			float64(dev.Present),
			dev.Identifier,
			dev.Productname,
			dev.Name,
		)

		if dev.CanMeasureTemp() {
			err = stringToFloatMetric(
				ch, fc.TemperatureDesc, dev.Temperature.FmtCelsius(),
				dev.Identifier, dev.Productname, dev.Name,
			)
			if err != nil {
				log.Printf("Unable to parse temperature data of \"%s\" : %v\n", dev.Name, err)
			}

			err = stringToFloatMetric(
				ch, fc.TemperatureOffsetDesc, dev.Temperature.FmtOffset(),
				dev.Identifier, dev.Productname, dev.Name,
			)
			if err != nil {
				log.Printf("Unable to parse temperature offset data of \"%s\" : %v\n", dev.Name, err)
			}

		}

		if dev.CanMeasurePower() {
			err = stringToFloatMetric(
				ch, fc.EnergyWhDesc, dev.Powermeter.FmtEnergyWh(),
				dev.Identifier, dev.Productname, dev.Name,
			)
			if err != nil {
				log.Printf("Unable to parse energy data of \"%s\" : %v\n", dev.Name, err)
			}

			err = stringToFloatMetric(
				ch, fc.PowerWDesc, dev.Powermeter.FmtPowerW(),
				dev.Identifier, dev.Productname, dev.Name,
			)
			if err != nil {
				log.Printf("Unable to parse power data of \"%s\" : %v\n", dev.Name, err)
			}
		}

		if dev.IsSwitch() {
			ch <- prometheus.MustNewConstMetric(
				fc.SwitchState,
				prometheus.GaugeValue,
				parseSwitchStrings(dev.Switch.State),
				dev.Identifier,
				dev.Productname,
				dev.Name,
			)
			ch <- prometheus.MustNewConstMetric(
				fc.SwitchMode,
				prometheus.GaugeValue,
				parseSwitchStrings(dev.Switch.Mode),
				dev.Identifier,
				dev.Productname,
				dev.Name,
			)
			ch <- prometheus.MustNewConstMetric(
				fc.SwitchBoxLock,
				prometheus.GaugeValue,
				parseSwitchStrings(dev.Switch.Lock),
				dev.Identifier,
				dev.Productname,
				dev.Name,
			)
			ch <- prometheus.MustNewConstMetric(
				fc.SwitchDeviceLock,
				prometheus.GaugeValue,
				parseSwitchStrings(dev.Switch.DeviceLock),
				dev.Identifier,
				dev.Productname,
				dev.Name,
			)
		}
	}
}

func NewFritzCollector() *fritzCollector {
	return &fritzCollector{
		InfoDesc: prometheus.NewDesc(
			"fritzbox_device_info",
			"Device information",
			append(genericLabels,
				"internal_id", "fw_version", "manufacturer", "functionbitmask",
			),
			prometheus.Labels{},
		),
		PresentDesc: prometheus.NewDesc(
			"fritzbox_device_present",
			"Device connected (1) or not (0)",
			genericLabels,
			prometheus.Labels{},
		),
		TemperatureDesc: prometheus.NewDesc(
			"fritzbox_temperature",
			"Temperature measured at the device sensor in units of 0.1 °C",
			genericLabels,
			prometheus.Labels{},
		),
		TemperatureOffsetDesc: prometheus.NewDesc(
			"fritzbox_temperature_offset",
			"Temperature offset (set by the user) in units of 0.1 °C",
			genericLabels,
			prometheus.Labels{},
		),
		EnergyWhDesc: prometheus.NewDesc(
			"fritzbox_energy",
			"Absolute energy consumption since the device started operating",
			genericLabels,
			prometheus.Labels{},
		),
		PowerWDesc: prometheus.NewDesc(
			"fritzbox_power",
			"Current power, refreshed approx every 2 minutes",
			genericLabels,
			prometheus.Labels{},
		),
		SwitchState: prometheus.NewDesc(
			"fritzbox_switch_state",
			"Switch state 1/0 (on/off), -1 if not known or error",
			genericLabels,
			prometheus.Labels{},
		),
		SwitchMode: prometheus.NewDesc(
			"fritzbox_switch_mode",
			"Switch mode 1/0 (manual/automatic), -1 if not known or error",
			genericLabels,
			prometheus.Labels{},
		),
		SwitchBoxLock: prometheus.NewDesc(
			"fritzbox_switch_boxlock",
			"Switching via FRITZ!Box disabled? 1/0, -1 if not known or error",
			genericLabels,
			prometheus.Labels{},
		),
		SwitchDeviceLock: prometheus.NewDesc(
			"fritzbox_switch_devicelock",
			"Switching via device disabled 1/0, -1 if not known or error",
			genericLabels,
			prometheus.Labels{},
		),
	}
}

// stringToFloatMetric converts a string `val` into a valid float metric
func stringToFloatMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, val, identifier, productName, name string) error {
	tc, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		tc,
		identifier,
		productName,
		name,
	)
	return nil
}

// parseSwitchStrings parses state strings of switches into floats
func parseSwitchStrings(val string) float64 {
	switch val {
	case "0", "automatic":
		return 0.0
	case "1", "manual":
		return 1.0
	default:
		return -1.0
	}
}
