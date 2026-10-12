# hardware-events

[![Build](https://github.com/creativeprojects/hardware-events/actions/workflows/build.yml/badge.svg)](https://github.com/creativeprojects/hardware-events/actions/workflows/build.yml)

A small daemon that keeps a home server quiet and cool: it reads temperature sensors, drives the fans through `ipmitool` (or any other command), spins down idle hard drives, and exposes everything to Prometheus or Zabbix.

The project started when I changed the fans on a Supermicro motherboard and they started to run at full speed, because their nominal speed was measured too low for the BMC. I could have bought other fans, but I didn't want to. I have it running on various Supermicro server boards (X9/X10/X11) and manage the different fan zones separately.

## Features

* Reads temperatures from `sysfs` files or from the output of any command (CPU, chipset, SATA and NVMe drives, `smartctl`, ...).
* Controls fan speed per zone with linear rules (temperature range → fan speed range), with a moving average to avoid oscillation.
* Several sensors can drive the same zone: the zone always runs at the highest speed requested by any of its sensors.
* Hands fan control back to the motherboard when the daemon stops.
* Monitors disk activity from `/proc/diskstats` and puts idle drives in standby, while still reporting their temperature when they are awake.
* Exposes `disk_temperature`, `disk_active` and `fan_speed` as Prometheus metrics.
* Runs scheduled tasks fed by Go templates, for example `zabbix_sender`.
* Integrates with systemd (`Type=notify` and watchdog).
* Simulation mode to test your rules without touching the hardware.

## Requirements

* Linux. FreeBSD binaries are built but not tested, and you will need to adapt the sensor paths since the defaults rely on Linux `sysfs`.
* Root privileges (or equivalent) to run `ipmitool`, `hdparm` and read some sensors.
* The command line tools you reference in the configuration, typically `ipmitool`, `hdparm` and optionally `smartctl`.
* For drive temperatures via `/sys/block/*/device/hwmon`, a kernel with the `drivetemp` module (5.6+).

## Installation

### Packages

Each [release](https://github.com/creativeprojects/hardware-events/releases) provides `.deb`, `.rpm` and `.apk` packages for `amd64`, `386`, `arm` (v6/v7) and `arm64`, plus `tar.gz` archives for Linux and FreeBSD.

The packages install:

| Path | Purpose |
|------|---------|
| `/opt/hardware-events/hardware-events` | binary |
| `/opt/hardware-events/config.yaml` | configuration (kept on upgrade) |
| `/opt/hardware-events/zabbix_template.go.txt` | Zabbix template (kept on upgrade) |
| `/etc/systemd/system/hardware-events.service` | systemd unit |

The post-install script enables and starts the service. Edit `config.yaml` for your hardware, then `systemctl restart hardware-events`.

### From source

```sh
git clone https://github.com/creativeprojects/hardware-events.git
cd hardware-events
make build          # builds ./hardware-events
make test           # runs the test suite
```

`make deploy` copies the binary, `config-local.yaml`, the templates and the unit file to `/opt/hardware-events/`. `make deploy-<host>` does the same over `rsync` using `config-<host>.yaml`.

## Running

```
hardware-events [-c config.yaml] [-q | -v | -d] [-s [-r1 seed] [-r2 seed]]
```

| Flag | Description |
|------|-------------|
| `-c` | configuration file (default `config.yaml`) |
| `-q` | quiet: only warnings and errors |
| `-v` | verbose: debug output |
| `-d` | debug: trace output, including every sensor reading |
| `-s` | simulation mode (see below) |
| `-r1`, `-r2` | random seeds used in simulation mode (default 42) |

A relative configuration path is resolved **from the directory of the binary**, not from the current directory. The same applies to template `source` paths.

The daemon runs until it receives `SIGINT` or `SIGTERM`. On shutdown it runs the fan `exit_command` so the motherboard takes control of the fans again.

### systemd

The provided unit uses `Type=notify` with a 15 minute watchdog. The daemon notifies systemd when it is ready, when it is stopping, and pings the watchdog periodically.

```sh
systemctl enable --now hardware-events
journalctl -u hardware-events -f
```

## How fan control works

Each fan **zone** (for example CPU fans and chassis fans) has an ID understood by your `set_command` and a list of **sensors** to watch. Each sensor has a list of **rules**:

1. On every tick (`run_every`), the sensor temperature is read and added to a moving average covering the `average` duration.
2. The first rule whose `temperature` range contains the averaged value is applied. The fan speed is interpolated linearly between `fan_speed.from` at `temperature.from` and `fan_speed.to` at `temperature.to`. Alternatively, a rule can request a fixed speed with `fan_speed.set`.
3. If no rule matches, the sensor requests the zone `min_speed` when below all the rules, or `max_speed` when above.
4. A rule can carry its own `run_every` to read the sensor more often while that rule is active. The default interval is restored when the rule no longer matches.
5. The zone collects the speed requested by each of its sensors and sets the fan to the **highest** one, clamped between `min_speed` and `max_speed`. The command is only run when the speed changes.

## Configuration

The configuration is a single YAML file. A complete example is in [config.yaml](config.yaml); the sections below describe each key.

Commands are run through `/bin/sh -c`. Some placeholders are expanded before running a command or opening a file:

| Placeholder | Available in | Value |
|-------------|--------------|-------|
| `${DEVICE}` | disk sensors, `disk_power_status` | resolved device path, e.g. `/dev/sda` |
| `${DEVICE_NAME}` | disk sensors, `disk_power_status` | kernel device name, e.g. `sda` |
| `${FAN_ZONE}` | `fan_control.set_command` | zone `id` |
| `${FAN_SPEED}` | `fan_control.set_command` | speed in percent |

### `sensors`

Named sources of integer temperature readings. A sensor is one of:

```yaml
sensors:
  # run a command and extract the value with a regular expression (first capture group)
  smartctl:
    command: "/usr/sbin/smartctl -l scttempsts ${DEVICE}"
    regexp: "Current Temperature:\\s+(\\d+) Celsius"
    timeout: 5s

  # read a single file; globs are allowed but must match exactly one file
  cpu:
    file: "/sys/devices/platform/coretemp.0/hwmon/hwmon*/temp1_input"
    divider: 1000

  # read several files and average the values
  cpu-average:
    files:
    - "/sys/devices/platform/coretemp.0/hwmon/hwmon*/temp?_input"
    divider: 1000

  pch:
    file: "/sys/class/thermal/thermal_zone1/temp"
    divider: 1000

  satatemp:
    file: "/sys/block/${DEVICE_NAME}/device/hwmon/hwmon*/temp1_input"
    divider: 1000

  nvmetemp:
    file: "/sys/block/${DEVICE_NAME}/device/hwmon*/temp1_input"
    divider: 1000
```

| Key | Description |
|-----|-------------|
| `command` | command to run; its standard output is the value unless `regexp` is set |
| `regexp` | regular expression applied to the command output; the first capture group is the value |
| `file` | path or glob of a file containing the value |
| `files` | list of paths or globs; all matching values are averaged |
| `divider` | divide the raw value (sysfs reports millidegrees, so use `1000`) |
| `timeout` | maximum duration for `command` |

Sensors that use `${DEVICE}` or `${DEVICE_NAME}` are meant to be referenced by a disk (see `disks`), which provides the device.

### `fan_control`

```yaml
fan_control:
  init_command: "/usr/bin/ipmitool raw 0x30 0x45 0x01 0x01"   # set fan mode to full
  set_command: "/usr/bin/ipmitool raw 0x30 0x70 0x66 0x01 ${FAN_ZONE} ${FAN_SPEED}"
  exit_command: "/usr/bin/ipmitool raw 0x30 0x45 0x01 0x00"   # set fan mode back to normal
  timeout: 5s
  parameters:
    FAN_ZONE:
      format: "%#x"     # ipmitool wants hexadecimal
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
    zone2:
      id: 1
      min_speed: 25
      run_every: 5m
      sensors:
        datapool1:
          average: 5m
          rules:
          - temperature:
              from: 40
              to: 60
            fan_speed:
              from: 30
              to: 100
```

| Key | Description |
|-----|-------------|
| `init_command` | run once at startup, typically to take manual control of the fans (optional) |
| `set_command` | run each time a zone speed changes; use `${FAN_ZONE}` and `${FAN_SPEED}` |
| `exit_command` | run at shutdown to give control back to the motherboard (optional) |
| `timeout` | maximum duration for each command |
| `parameters.<NAME>.format` | `fmt` format applied to a placeholder before expansion (default `%d`) |

Zone keys:

| Key | Description |
|-----|-------------|
| `id` | zone identifier passed as `${FAN_ZONE}` |
| `min_speed` | lowest speed ever set, in percent (default 25) |
| `max_speed` | highest speed ever set (default 100) |
| `run_every` | default read interval for the sensors of this zone |
| `sensors` | map of sensor name → sensor settings. The name is either a key of `sensors` or a key of `disks` |

Sensor keys inside a zone:

| Key | Description |
|-----|-------------|
| `average` | length of the moving average window. Must be at least `run_every` |
| `run_every` | read interval, overrides the zone value |
| `rules` | list of rules, see below |

Rule keys:

| Key | Description |
|-----|-------------|
| `temperature.from`, `temperature.to` | temperature range the rule applies to, in °C |
| `fan_speed.from`, `fan_speed.to` | speed at `temperature.from` and `temperature.to`, interpolated in between |
| `fan_speed.set` | fixed speed for the whole range, instead of `from`/`to` |
| `run_every` | read interval while this rule is active (optional) |

### `disk_power_status`

Named methods to check whether a drive is spinning and to put it in standby. The check output is matched against the `active`, `standby` and `sleeping` strings.

```yaml
disk_power_status:
  hdparm:
    check_command: "/sbin/hdparm -C ${DEVICE}"
    active: "active/idle"
    standby: "standby"
    sleeping: "sleeping"
    standby_command: "/sbin/hdparm -y ${DEVICE}"
    timeout: 5s
```

`file` can be used instead of `check_command` to read the status from a `sysfs` file. When a single method is defined, every disk uses it; otherwise set `power_status` on each disk.

### `disks`

```yaml
disks:
  rpool1:
    device: "/dev/disk/by-id/ata-SAMSUNG_SSD_first"
    temperature_sensor: satatemp
    monitor_temperature: always
  datapool1:
    device: "/dev/disk/by-id/ata-ST2000DM001-first"
    temperature_sensor: satatemp
    monitor_temperature: when_active
    last_active: 50m
    standby_after: 1h
```

| Key | Description |
|-----|-------------|
| `device` | device path. Symlinks such as `/dev/disk/by-id/...` are resolved to the kernel device |
| `temperature_sensor` | name of a sensor from `sensors` to read this disk temperature |
| `monitor_temperature` | `always`, `when_active` (only when spinning and not idle, so a sleeping drive is never woken up) or `never` |
| `last_active` | how long without I/O before the disk is considered idle (default 1m) |
| `standby_after` | put the disk in standby after this much time without I/O. Leave unset to never force standby |
| `check_every` | how often the standby watch checks the disk (default 5m, with ±1 minute jitter) |
| `power_status` | name of a `disk_power_status` method, needed when more than one is defined |

Disk activity comes from `/proc/diskstats` (read and write sectors on the disk partitions). Disk status and temperature readings are cached for one minute.

### `disk_pools`

Groups disks by name, for reporting only. The pool name is attached to the disk metrics and the Zabbix template reports the number of active disks per pool.

```yaml
disk_pools:
  rpool:
    - rpool1
    - rpool2
  datapool:
    - datapool1
    - datapool2
```

### `templates`, `tasks` and `schedule`

Tasks run a command, optionally with a rendered [Go template](https://pkg.go.dev/text/template) on its standard input. Schedules decide when a task runs: `startup` and/or `every <duration>`.

```yaml
templates:
  zabbix:
    source: "zabbix_template.go.txt"   # relative to the binary

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

The template receives the whole runtime state. The most useful fields are `.Disks` (with `.Name`, `.Pool`, `.IsActive`, `.HasTemperature`, `.TemperatureAvailable`, `.Temperature`) and `.DiskPools` (with `.Name`, `.CountActive`). See [zabbix_template.go.txt](zabbix_template.go.txt) for an example.

### `telemetry`

```yaml
telemetry:
  prometheus:
    enabled: true
    listen_address: ":8080"
```

When enabled, metrics are served at `http://<listen_address>/metrics`.

## Prometheus metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `disk_temperature` | `name`, `device`, `pool` | disk temperature in °C, only reported when the temperature is available |
| `disk_active` | `name`, `device`, `pool` | 1 when the disk is spinning, 0 otherwise. `pool` is only set for disks that belong to a pool |
| `fan_speed` | `name`, `id` | current zone fan speed in percent |

## Zabbix

The bundled template feeds `zabbix_sender` with:

* `disk.pools` and `disk.names`: JSON arrays for low-level discovery (`{#POOL_NAME}`, `{#DISK_NAME}`)
* `disks.active[<pool>]`: number of active disks in the pool
* `disks.temperature[<disk>]`: disk temperature, or `0` when the disk is in standby

## Simulation mode

Start with `-s` (or set `simulation: true` in the configuration) to try a configuration on a machine without the hardware. Sensors return random temperatures driven by the `-r1`/`-r2` seeds, disks start active and go to standby when asked, and every command is logged instead of being run. Use `-d` to see the readings and the resulting fan speeds.

```sh
hardware-events -s -d -c config.yaml
```

## External resources

* Supermicro fan control through IPMI raw commands: https://forums.servethehome.com/index.php?resources/supermicro-x9-x10-x11-fan-speed-control.20/

## License

[GPL v3](LICENSE)
