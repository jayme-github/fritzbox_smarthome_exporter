# FRITZ!Box Smarthome exporter for prometheus
Export information about your smarthome devices (switches, powermeters, thermostat, ...) to prometheus.

[![Docker Pulls](https://img.shields.io/docker/pulls/jaymedh/fritzbox_smarthome_exporter)](https://hub.docker.com/repository/docker/jaymedh/fritzbox_smarthome_exporter)

# Usage

* Setup a (restricted) user account for the exporter to use. This accounts only need access to smarthome devices, see [Setting up users with restricted authorization](https://en.avm.de/service/fritzbox/fritzbox-5490/knowledge-base/publication/show/1522_Accessing-FRITZ-Box-from-the-home-network-with-user-accounts/) for details.
* [Download your FRITZ!Box certificate](https://en.avm.de/service/fritzbox/fritzbox-7390/knowledge-base/publication/show/1523_Downloading-your-FRITZ-Box-certificate-and-importing-it-to-your-computer/) (recommended)

```
Usage of ./fritzbox_smarthome_exporter:
  -cert="": Path to the FRITZ!Box certificate.
  -noverify=false: Omit TLS verification of the FRITZ!Box certificate.
  -password="": FRITZ!Box password.
  -url="https://fritz.box": FRITZ!Box URL.
  -username="": FRITZ!Box username.
```
Command line arguments or environment variables (the argument as uppercase, like `CERT` for `-cert`) may be used.


The exporter will bind to TCP Port 9103 and export the following metrics via `/metrics`:

```
# HELP fritzbox_device_info Device information
# TYPE fritzbox_device_info gauge
fritzbox_device_info{device_id="01111 0111111",device_name="Switch 1",device_type="FRITZ!DECT 200",functionbitmask="2944",fw_version="03.87",internal_id="16",manufacturer="AVM"} 1
fritzbox_device_info{device_id="12345 0000000",device_name="HKR 1",device_type="Comet DECT",functionbitmask="320",fw_version="03.68",internal_id="23",manufacturer="AVM"} 1
# HELP fritzbox_device_present Device connected (1) or not (0)
# TYPE fritzbox_device_present gauge
fritzbox_device_present{device_id="01111 0111111",device_name="Switch 1",device_type="FRITZ!DECT 200"} 1
fritzbox_device_present{device_id="12345 0000000",device_name="HKR 1",device_type="Comet DECT"} 1
# HELP fritzbox_energy Absolute energy consumption since the device started operating
# TYPE fritzbox_energy gauge
fritzbox_energy{device_id="01111 0111111",device_name="Switch 1",device_type="FRITZ!DECT 200"} 103023
# HELP fritzbox_power Current power, refreshed approx every 2 minutes
# TYPE fritzbox_power gauge
fritzbox_power{device_id="01111 0111111",device_name="Switch 1",device_type="FRITZ!DECT 200"} 0
# HELP fritzbox_switch_boxlock Switching via FRITZ!Box disabled? 1/0, -1 if not known or error
# TYPE fritzbox_switch_boxlock gauge
fritzbox_switch_boxlock{device_id="01111 0111111",device_name="Switch 1",device_type="FRITZ!DECT 200"} 0
# HELP fritzbox_switch_devicelock Switching via device disabled 1/0, -1 if not known or error
# TYPE fritzbox_switch_devicelock gauge
fritzbox_switch_devicelock{device_id="01111 0111111",device_name="Switch 1",device_type="FRITZ!DECT 200"} 1
# HELP fritzbox_switch_mode Switch mode 1/0 (manual/automatic), -1 if not known or error
# TYPE fritzbox_switch_mode gauge
fritzbox_switch_mode{device_id="01111 0111111",device_name="Switch 1",device_type="FRITZ!DECT 200"} -1
# HELP fritzbox_switch_state Switch state 1/0 (on/off), -1 if not known or error
# TYPE fritzbox_switch_state gauge
fritzbox_switch_state{device_id="01111 0111111",device_name="Switch 1",device_type="FRITZ!DECT 200"} -1
# HELP fritzbox_temperature Temperature measured at the device sensor in units of 0.1 °C
# TYPE fritzbox_temperature gauge
fritzbox_temperature{device_id="01111 0111111",device_name="Switch 1",device_type="FRITZ!DECT 200"} 19
fritzbox_temperature{device_id="12345 0000000",device_name="HKR 1",device_type="Comet DECT"} 22
# HELP fritzbox_temperature_offset Temperature offset (set by the user) in units of 0.1 °C
# TYPE fritzbox_temperature_offset gauge
fritzbox_temperature_offset{device_id="01111 0111111",device_name="Switch 1",device_type="FRITZ!DECT 200"} -1
fritzbox_temperature_offset{device_id="12345 0000000",device_name="HKR 1",device_type="Comet DECT"} -0.5
# HELP fritzbox_thermostat_battery_charge_level Battery charge level in percent
# TYPE fritzbox_thermostat_battery_charge_level gauge
fritzbox_thermostat_battery_charge_level{device_id="44363 2777777",device_name="HKR_1",device_type="Comet DECT"} 70
# HELP fritzbox_thermostat_batterylow 0 if the battery is OK, 1 if it is running low on capacity (this seems to be very unreliable)
# TYPE fritzbox_thermostat_batterylow gauge
fritzbox_thermostat_batterylow{device_id="44363 2777777",device_name="HKR_1",device_type="Comet DECT"} 0
# HELP fritzbox_thermostat_comfort Comfort temperature configured in units of 0.1 °C
# TYPE fritzbox_thermostat_comfort gauge
fritzbox_thermostat_comfort{device_id="44363 2777777",device_name="HKR_1",device_type="Comet DECT"} 19
# HELP fritzbox_thermostat_errorcode Thermostat error code (0 = OK), see https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf
# TYPE fritzbox_thermostat_errorcode gauge
fritzbox_thermostat_errorcode{device_id="44363 2777777",device_name="HKR_1",device_type="Comet DECT"} 0
# HELP fritzbox_thermostat_goal Desired temperature (user controlled) in units of 0.1 °C
# TYPE fritzbox_thermostat_goal gauge
fritzbox_thermostat_goal{device_id="44363 2777777",device_name="HKR_1",device_type="Comet DECT"} 17
# HELP fritzbox_thermostat_saving Configured energy saving temperature in units of 0.1 °C
# TYPE fritzbox_thermostat_saving gauge
fritzbox_thermostat_saving{device_id="44363 2777777",device_name="HKR_1",device_type="Comet DECT"} 16
# HELP fritzbox_thermostat_window_open 1 if detected an open window (usually turns off heating), 0 if not.
# TYPE fritzbox_thermostat_window_open gauge
fritzbox_thermostat_window_open{device_id="44363 2777777",device_name="HKR_1",device_type="Comet DECT"} 0
```


# Docker
Docker images are build for tags [jaymedh/fritzbox_smarthome_exporter](https://hub.docker.com/r/jaymedh/fritzbox_smarthome_exporter/).

FRITZ!Box certificate may be mounted into the container, configuration can be done via arguments or environment variables (or both):
```
docker run -d --name fritzbox_smarthome_exporter -p 9103:9103 \
  -v $(pwd)/boxcert.cer:/fritzbox.pem:ro \
  -e PASSWORD=SuperSecret \
  -e USERNAME=SmarthomeUser \
  jaymedh/fritzbox_smarthome_exporter:v0.0.1 -url="https://fritz.box:8443" -cert=/fritzbox.pem
```


# Grafana

Example Grafana dashboard can be found at https://grafana.com/dashboards/7019
