# hardware-events

This software helps with the monitoring of a server motherboard.

The project started when I changed the FANs on a Supermicro motherboard and they started to run at full speed because their nominal speed was measured too low for the motherboard. I could have bought some other FANs, but I didn't want to.

This project:
* reads temperature from various sensors (CPU, motherboard, disks)
* sets FAN speed via the command line like `ipmitool` (multi-zones)
* sends sensor and FAN data to a monitoring platform (zabbix)

I have it running on various Supermicro server boards (X9/X10/X11) and manage the different FAN zones separately.

Everything is in the configuration file:

```yaml
---
sensors:
  hddtemp:
    command: "/usr/sbin/hddtemp -n ${DEVICE}"
    timeout: 5s
  smartctl:
    command: "/usr/sbin/smartctl -l scttempsts ${DEVICE}"
    regexp: "Current Temperature:\\s+(\\d+) Celsius"
    timeout: 5s
  cpu:
    file: "/sys/devices/platform/coretemp.0/hwmon/hwmon*/temp1_input"
    divider: 1000
  cpu-average:
    files:
    - "/sys/devices/platform/coretemp.0/hwmon/hwmon*/temp?_input"
  pch:
    file: "/sys/class/thermal/thermal_zone1/temp"
    divider: 1000
  drivetemp:
    file: "/sys/block/${DEVICE_NAME}/device/hwmon/hwmon*/temp1_input"
    divider: 1000

fan_control:
  # set fan mode to full
  init_command: "/usr/bin/ipmitool raw 0x30 0x45 0x01 0x01"
  set_command: "/usr/bin/ipmitool raw 0x30 0x70 0x66 0x01 ${FAN_ZONE} ${FAN_SPEED}"
  # set fan mode to normal
  exit_command: "/usr/bin/ipmitool raw 0x30 0x45 0x01 0x00"
  timeout: 5s
  parameters:
    FAN_ZONE:
      format: "%#x"
    FAN_SPEED:
      format: "%#x"
  zones:
    zone1:
      id: 0
      min_speed: 25
      sensors:
        cpu:
          average: 30s
          run_every: 10s
          rules:
          - temperature:
              from: 40
              to: 60
            run_every: 5s
            fan_speed:
              from: 25
              to: 100
        pch:
          average: 1m
          run_every: 20s
          rules:
          - temperature:
              from: 50
              to: 70
            fan_speed:
              from: 25
              to: 100
    zone2:
      id: 1
      min_speed: 25
      run_every: 5m
      sensors:
        rpool1:
          average: 1m
          run_every: 1m
          rules:
          - temperature:
              from: 40
              to: 60
            fan_speed:
              from: 30
              to: 100
        datapool1:
          average: 5m
          rules:
          - temperature:
              from: 40
              to: 60
            fan_speed:
              from: 30
              to: 100
        datapool2:
          average: 5m
          rules:
          - temperature:
              from: 40
              to: 60
            fan_speed:
              from: 30
              to: 100
        datapool3:
          average: 5m
          rules:
          - temperature:
              from: 40
              to: 60
            fan_speed:
              from: 30
              to: 100
        datapool4:
          average: 5m
          rules:
          - temperature:
              from: 40
              to: 60
            fan_speed:
              from: 30
              to: 100

disk_power_status:
  check_command: "/sbin/hdparm -C ${DEVICE}"
  active: "active/idle"
  standby: "standby"
  sleeping: "sleeping"
  standby_command: "/sbin/hdparm -y ${DEVICE}"
  timeout: 5s

disk_pools:
  rpool:
    - rpool1
    - rpool2
  datapool:
    - datapool1
    - datapool2
    - datapool3
    - datapool4

disks:
  rpool1:
    device: "/dev/disk/by-id/ata-SAMSUNG_SSD_first"
    temperature_sensor: drivetemp
    monitor_temperature: always
  rpool2:
    device: "/dev/disk/by-id/ata-SAMSUNG_SSD_second"
    temperature_sensor: drivetemp
    monitor_temperature: always
  datapool1:
    device: "/dev/disk/by-id/ata-ST2000DM001-first"
    temperature_sensor: drivetemp
    monitor_temperature: when_active
    last_active: 50m
    standby_after: 1h
  datapool2:
    device: "/dev/disk/by-id/ata-ST2000DM001-second"
    temperature_sensor: drivetemp
    monitor_temperature: when_active
    last_active: 50m
    standby_after: 1h
  datapool3:
    device: "/dev/disk/by-id/ata-ST2000DM001-third"
    temperature_sensor: drivetemp
    monitor_temperature: when_active
    last_active: 50m
    standby_after: 1h
  datapool4:
    device: "/dev/disk/by-id/ata-ST2000DM001-fourth"
    temperature_sensor: drivetemp
    monitor_temperature: when_active
    last_active: 50m
    standby_after: 1h

templates:
  zabbix:
    source: "zabbix_template.go.txt"

tasks:
  zabbix_sender:
    command: "zabbix_sender -z 127.0.0.1 -s \"Zabbix server\" -i -"
    timeout: 5s
    stdin:
      template: zabbix

schedule:
  zabbix:
    task: zabbix_sender
    when:
    - startup
    - every 5m
```
