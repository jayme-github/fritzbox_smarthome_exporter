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
	Info                         *prometheus.Desc
	Present                      *prometheus.Desc
	Temperature                  *prometheus.Desc
	TemperatureOffset            *prometheus.Desc
	EnergyWh                     *prometheus.Desc
	PowerW                       *prometheus.Desc
	SwitchState                  *prometheus.Desc
	SwitchMode                   *prometheus.Desc
	SwitchBoxLock                *prometheus.Desc
	SwitchDeviceLock             *prometheus.Desc
	ThermostatBatteryChargeLevel *prometheus.Desc
	ThermostatBatteryLow         *prometheus.Desc
	ThermostatErrorCode          *prometheus.Desc
	ThermostatTempComfort        *prometheus.Desc
	ThermostatTempGoal           *prometheus.Desc
	ThermostatTempSaving         *prometheus.Desc
	ThermostatWindowOpen         *prometheus.Desc
}

func (fc *fritzCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- fc.Info
	ch <- fc.Present
	ch <- fc.Temperature
	ch <- fc.TemperatureOffset
	ch <- fc.EnergyWh
	ch <- fc.PowerW
	ch <- fc.SwitchState
	ch <- fc.SwitchMode
	ch <- fc.SwitchBoxLock
	ch <- fc.SwitchDeviceLock
	ch <- fc.ThermostatBatteryChargeLevel
	ch <- fc.ThermostatBatteryLow
	ch <- fc.ThermostatErrorCode
	ch <- fc.ThermostatTempComfort
	ch <- fc.ThermostatTempGoal
	ch <- fc.ThermostatTempSaving
	ch <- fc.ThermostatWindowOpen
}

func (fc *fritzCollector) Collect(ch chan<- prometheus.Metric) {
	var err error
	l, err := fritzClient.SafeList()

	if err != nil {
		log.Println("Unable to collect data:", err)
		ch <- prometheus.NewInvalidMetric(fc.Info, err)
		ch <- prometheus.NewInvalidMetric(fc.Present, err)
		ch <- prometheus.NewInvalidMetric(fc.Temperature, err)
		ch <- prometheus.NewInvalidMetric(fc.TemperatureOffset, err)
		ch <- prometheus.NewInvalidMetric(fc.EnergyWh, err)
		ch <- prometheus.NewInvalidMetric(fc.PowerW, err)
		ch <- prometheus.NewInvalidMetric(fc.SwitchState, err)
		ch <- prometheus.NewInvalidMetric(fc.SwitchMode, err)
		ch <- prometheus.NewInvalidMetric(fc.SwitchBoxLock, err)
		ch <- prometheus.NewInvalidMetric(fc.SwitchDeviceLock, err)
		ch <- prometheus.NewInvalidMetric(fc.ThermostatBatteryChargeLevel, err)
		ch <- prometheus.NewInvalidMetric(fc.ThermostatBatteryLow, err)
		ch <- prometheus.NewInvalidMetric(fc.ThermostatErrorCode, err)
		ch <- prometheus.NewInvalidMetric(fc.ThermostatTempComfort, err)
		ch <- prometheus.NewInvalidMetric(fc.ThermostatTempGoal, err)
		ch <- prometheus.NewInvalidMetric(fc.ThermostatTempSaving, err)
		ch <- prometheus.NewInvalidMetric(fc.ThermostatWindowOpen, err)
		return
	}

	for _, dev := range l.Devices {
		ch <- prometheus.MustNewConstMetric(
			fc.Info,
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
			fc.Present,
			prometheus.GaugeValue,
			float64(dev.Present),
			dev.Identifier,
			dev.Productname,
			dev.Name,
		)

		if dev.Present == 1 && dev.CanMeasureTemp() {
			if err := mustStringToFloatMetric(ch, fc.Temperature, dev.Temperature.FmtCelsius(), &dev); err != nil {
				log.Printf("Unable to parse temperature data of \"%s\" : %v\n", dev.Name, err)
			}

			if err := mustStringToFloatMetric(ch, fc.TemperatureOffset, dev.Temperature.FmtOffset(), &dev); err != nil {
				log.Printf("Unable to parse temperature offset data of \"%s\" : %v\n", dev.Name, err)
			}
		}

		if dev.Present == 1 && dev.CanMeasurePower() {
			if err := mustStringToFloatMetric(ch, fc.EnergyWh, dev.Powermeter.FmtEnergyWh(), &dev); err != nil {
				log.Printf("Unable to parse energy data of \"%s\" : %v\n", dev.Name, err)
			}

			if err := mustStringToFloatMetric(ch, fc.PowerW, dev.Powermeter.FmtPowerW(), &dev); err != nil {
				log.Printf("Unable to parse power data of \"%s\" : %v\n", dev.Name, err)
			}
		}

		if dev.IsThermostat() {
			// Battery charge level is optional
			if err := canStringToFloatMetric(ch, fc.ThermostatBatteryChargeLevel, dev.Thermostat.BatteryChargeLevel, &dev); err != nil {
				log.Printf("Unable to parse battery charge level of \"%s\" : %v\n", dev.Name, err)
			}

			if err := mustStringToFloatMetric(ch, fc.ThermostatBatteryLow, dev.Thermostat.BatteryLow, &dev); err != nil {
				log.Printf("Unable to parse battery low state of \"%s\" : %v\n", dev.Name, err)
			}

			// Handle no error like error code 0
			errCodeStr := dev.Thermostat.ErrorCode
			if errCodeStr == "" {
				errCodeStr = "0"
			}
			if err := mustStringToFloatMetric(ch, fc.ThermostatErrorCode, errCodeStr, &dev); err != nil {
				log.Printf("Unable to parse thermostat error code of \"%s\" : %v\n", dev.Name, err)
			}

			// Comfort, Goal and Saving temperature are optional
			if err := canStringToFloatMetric(ch, fc.ThermostatTempComfort, dev.Thermostat.FmtComfortTemperature(), &dev); err != nil {
				log.Printf("Unable to parse comfort temperature of \"%s\" : %v\n", dev.Name, err)
			}
			if err := canStringToFloatMetric(ch, fc.ThermostatTempGoal, dev.Thermostat.FmtGoalTemperature(), &dev); err != nil {
				log.Printf("Unable to parse goal temperature of \"%s\" : %v\n", dev.Name, err)
			}
			if err := canStringToFloatMetric(ch, fc.ThermostatTempSaving, dev.Thermostat.FmtSavingTemperature(), &dev); err != nil {
				log.Printf("Unable to parse saving temperature of \"%s\" : %v\n", dev.Name, err)
			}

			// Window Open is optional
			if err := canStringToFloatMetric(ch, fc.ThermostatWindowOpen, dev.Thermostat.WindowOpen, &dev); err != nil {
				log.Printf("Unable to parse window open state of \"%s\" : %v\n", dev.Name, err)
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
		Info: prometheus.NewDesc(
			"fritzbox_device_info",
			"Device information",
			append(genericLabels,
				"internal_id", "fw_version", "manufacturer", "functionbitmask",
			),
			prometheus.Labels{},
		),
		Present: prometheus.NewDesc(
			"fritzbox_device_present",
			"Device connected (1) or not (0)",
			genericLabels,
			prometheus.Labels{},
		),
		Temperature: prometheus.NewDesc(
			"fritzbox_temperature",
			"Temperature measured at the device sensor in units of 0.1 °C",
			genericLabels,
			prometheus.Labels{},
		),
		TemperatureOffset: prometheus.NewDesc(
			"fritzbox_temperature_offset",
			"Temperature offset (set by the user) in units of 0.1 °C",
			genericLabels,
			prometheus.Labels{},
		),
		EnergyWh: prometheus.NewDesc(
			"fritzbox_energy",
			"Absolute energy consumption (in  Wh) since the device started operating",
			genericLabels,
			prometheus.Labels{},
		),
		PowerW: prometheus.NewDesc(
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
		ThermostatBatteryChargeLevel: prometheus.NewDesc(
			"fritzbox_thermostat_battery_charge_level",
			"Battery charge level in percent",
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
		ThermostatTempComfort: prometheus.NewDesc(
			"fritzbox_thermostat_comfort",
			"Comfort temperature configured in units of 0.1 °C",
			genericLabels,
			prometheus.Labels{},
		),
		ThermostatTempGoal: prometheus.NewDesc(
			"fritzbox_thermostat_goal",
			"Desired temperature (user controlled) in units of 0.1 °C",
			genericLabels,
			prometheus.Labels{},
		),
		ThermostatTempSaving: prometheus.NewDesc(
			"fritzbox_thermostat_saving",
			"Configured energy saving temperature in units of 0.1 °C",
			genericLabels,
			prometheus.Labels{},
		),
		ThermostatWindowOpen: prometheus.NewDesc(
			"fritzbox_thermostat_window_open",
			"1 if detected an open window (usually turns off heating), 0 if not.",
			genericLabels,
			prometheus.Labels{},
		),
	}
}

// stringToFloatMetric converts a string `val` into a valid float metric
func stringToFloatMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, value string, dev *fritz.Device, optional bool) error {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		if !optional {
			ch <- prometheus.NewInvalidMetric(desc, err)
		}
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
func canStringToFloatMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, value string, dev *fritz.Device) error {
	return stringToFloatMetric(ch, desc, value, dev, true)
}
func mustStringToFloatMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, value string, dev *fritz.Device) error {
	return stringToFloatMetric(ch, desc, value, dev, false)
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
