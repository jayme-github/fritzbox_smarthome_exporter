package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/bpicode/fritzctl/fritz"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	genericLabels          = []string{"device_id", "device_type", "device_name"}
	ErrParsingSwitchString = errors.New("Error parsing switch string")
)

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
	ThermostatBatteryLow  *prometheus.Desc
	ThermostatErrorCode   *prometheus.Desc
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
	ch <- fc.ThermostatBatteryLow
	ch <- fc.ThermostatErrorCode
}

func (fc *fritzCollector) Collect(ch chan<- prometheus.Metric) {
	var err error
	l, err := fritzClient.SafeList()

	if err != nil {
		log.Println("Unable to collect data:", err)
		ch <- prometheus.NewInvalidMetric(fc.InfoDesc, err)
		ch <- prometheus.NewInvalidMetric(fc.PresentDesc, err)
		ch <- prometheus.NewInvalidMetric(fc.TemperatureDesc, err)
		ch <- prometheus.NewInvalidMetric(fc.TemperatureOffsetDesc, err)
		ch <- prometheus.NewInvalidMetric(fc.EnergyWhDesc, err)
		ch <- prometheus.NewInvalidMetric(fc.PowerWDesc, err)
		ch <- prometheus.NewInvalidMetric(fc.SwitchState, err)
		ch <- prometheus.NewInvalidMetric(fc.SwitchMode, err)
		ch <- prometheus.NewInvalidMetric(fc.SwitchBoxLock, err)
		ch <- prometheus.NewInvalidMetric(fc.SwitchDeviceLock, err)
		ch <- prometheus.NewInvalidMetric(fc.ThermostatBatteryLow, err)
		ch <- prometheus.NewInvalidMetric(fc.ThermostatErrorCode, err)
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

		if dev.Present == 1 && dev.CanMeasureTemp() {
			if err := stringToFloatMetric(ch, fc.TemperatureDesc, dev.Temperature.FmtCelsius(), &dev); err != nil {
				log.Printf("Unable to parse temperature data of \"%s\" : %v\n", dev.Name, err)
			}

			if err := stringToFloatMetric(ch, fc.TemperatureOffsetDesc, dev.Temperature.FmtOffset(), &dev); err != nil {
				log.Printf("Unable to parse temperature offset data of \"%s\" : %v\n", dev.Name, err)
			}
		}

		if dev.Present == 1 && dev.CanMeasurePower() {
			if err := stringToFloatMetric(ch, fc.EnergyWhDesc, dev.Powermeter.FmtEnergyWh(), &dev); err != nil {
				log.Printf("Unable to parse energy data of \"%s\" : %v\n", dev.Name, err)
			}

			if err := stringToFloatMetric(ch, fc.PowerWDesc, dev.Powermeter.FmtPowerW(), &dev); err != nil {
				log.Printf("Unable to parse power data of \"%s\" : %v\n", dev.Name, err)
			}
		}

		if dev.IsThermostat() {
			if batteryLow, err := strconv.ParseFloat(dev.Thermostat.BatteryLow, 64); err != nil {
				ch <- prometheus.NewInvalidMetric(fc.ThermostatBatteryLow, err)
				log.Printf("Unable to parse battery low state of \"%s\" : %v\n", dev.Name, err)
			} else {
				ch <- prometheus.MustNewConstMetric(
					fc.ThermostatBatteryLow,
					prometheus.GaugeValue,
					batteryLow,
					dev.Identifier,
					dev.Productname,
					dev.Name,
				)
			}

			var errCode float64
			// Reset err so it can be used later to decide if we need to send the ThermostatErrCode metric
			err = nil
			if dev.Thermostat.ErrorCode != "" {
				errCode, err = strconv.ParseFloat(dev.Thermostat.ErrorCode, 64)
				if err != nil {
					ch <- prometheus.NewInvalidMetric(fc.ThermostatErrorCode, err)
					log.Printf("Unable to parse thermostat error code of \"%s\" : %v\n", dev.Name, err)
				}
			}
			if err == nil {
				ch <- prometheus.MustNewConstMetric(
					fc.ThermostatErrorCode,
					prometheus.GaugeValue,
					errCode,
					dev.Identifier,
					dev.Productname,
					dev.Name,
				)
			}
		}

		if dev.IsSwitch() {
			if err := switchMetric(ch, fc.SwitchState, dev.Switch.State, &dev); err != nil {
				log.Printf("Unable to parse switch state of \"%s\" : %v\n", dev.Name, err)
			}
			if err := switchMetric(ch, fc.SwitchMode, dev.Switch.Mode, &dev); err != nil {
				log.Printf("Unable to parse switch mode of \"%s\" : %v\n", dev.Name, err)
			}
			if err := switchMetric(ch, fc.SwitchBoxLock, dev.Switch.Lock, &dev); err != nil {
				log.Printf("Unable to parse switch lock of \"%s\" : %v\n", dev.Name, err)
			}
			if err := switchMetric(ch, fc.SwitchDeviceLock, dev.Switch.DeviceLock, &dev); err != nil {
				log.Printf("Unable to parse switch device lock of \"%s\" : %v\n", dev.Name, err)
			}
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
			"Absolute energy consumption (in  Wh) since the device started operating",
			genericLabels,
			prometheus.Labels{},
		),
		PowerWDesc: prometheus.NewDesc(
			"fritzbox_power",
			"Current power (in W), refreshed approx every 2 minutes",
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
		ThermostatBatteryLow: prometheus.NewDesc(
			"fritzbox_thermostat_batterylow",
			"0 if the battery is OK, 1 if it is running low on capacity (this seems to be very unreliable)",
			genericLabels,
			prometheus.Labels{},
		),
		ThermostatErrorCode: prometheus.NewDesc(
			"fritzbox_thermostat_errorcode",
			"Thermostat error code (0 = OK), see https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf",
			genericLabels,
			prometheus.Labels{},
		),
	}
}

// stringToFloatMetric converts a string `val` into a valid float metric
func stringToFloatMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, value string, dev *fritz.Device) error {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(desc, err)
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		val,
		dev.Identifier,
		dev.Productname,
		dev.Name,
	)
	return nil
}

// parseSwitchStrings parses state strings of switches into floats
func parseSwitchStrings(val string) (float64, error) {
	switch val {
	case "0", "auto":
		return 0.0, nil
	case "1", "manuell":
		return 1.0, nil
	default:
		return -1.0, fmt.Errorf("%s: %w", val, ErrParsingSwitchString)
	}
}

// switchMetrics creates switch metrics
func switchMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, value string, dev *fritz.Device) error {
	val, err := parseSwitchStrings(value)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(desc, err)
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		val,
		dev.Identifier,
		dev.Productname,
		dev.Name,
	)
	return nil
}
