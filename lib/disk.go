package lib

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/hardware-events/cache"
	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/lib/enum"
)

// Disk activity and status
type Disk struct {
	global        *Global
	config        cfg.Disk
	Name          string
	Device        string
	active        *cache.CacheValue[int]
	temperature   *cache.CacheValue[int]
	lastActivity  time.Time
	activityMutex sync.Mutex
	stats         *Diskstats
	idleAfter     time.Duration
	standbyAfter  time.Duration
	diskStatus    DiskStatuser
}

// NewDisk creates a new disk activity and status monitor
func NewDisk(global *Global, name string, config cfg.Disk, diskStatuses map[string]DiskStatuser) (*Disk, error) {
	device := config.Device
	fi, err := os.Lstat(device)
	if err != nil {
		return nil, err
	}

	var diskStatus DiskStatuser
	if config.PowerStatus != "" {
		diskStatus = diskStatuses[config.PowerStatus]
	}
	if diskStatus == nil && len(diskStatuses) == 1 {
		// load the only one instead
		for _, value := range diskStatuses {
			diskStatus = value
			break
		}
	}
	if diskStatus == nil {
		clog.Warningf("disk %q has no power status available", name)
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		// resolve symlink into kernel device name
		dest, err := os.Readlink(device)
		if err != nil {
			return nil, err
		}
		if filepath.IsAbs(dest) {
			device = dest
		} else {
			device = filepath.Join(filepath.Dir(device), dest)
		}
	}

	idleAfter := 1 * time.Minute
	if config.LastActive != "" {
		idleAfter, err = time.ParseDuration(config.LastActive)
		if err != nil {
			return nil, err
		}
	}

	var standbyAfter time.Duration
	if config.StandbyAfter != "" {
		standbyAfter, err = time.ParseDuration(config.StandbyAfter)
		if err != nil {
			return nil, err
		}
	}

	clog.Debugf("device %s: %s", name, device)
	return &Disk{
		global:        global,
		config:        config,
		Name:          name,
		Device:        device,
		active:        cache.NewCacheValue[int](1 * time.Minute),
		temperature:   cache.NewCacheValue[int](1 * time.Minute),
		idleAfter:     idleAfter,
		standbyAfter:  standbyAfter,
		diskStatus:    diskStatus,
		activityMutex: sync.Mutex{},
	}, nil
}

// IsActive returns true when the disk is not in standby or sleep mode
func (d *Disk) IsActive() bool {
	if d.diskStatus == nil {
		// shouldn't happen?
		return false
	}
	output, err := d.active.Get(func() (int, error) {
		return int(d.diskStatus.Get(d.expandEnv)), nil
	})
	if err != nil {
		return false
	}
	return output == int(enum.DiskStatusActive)
}

// HasTemperature indicates if the disk accepts temperature readings.
func (d *Disk) HasTemperature() bool {
	return d.config.MonitorTemperature != "" && d.config.MonitorTemperature != "never"
}

// TemperatureAvailable indicates if we can read the disk temperature now.
// In general it means the disk is not idle
func (d *Disk) TemperatureAvailable() bool {
	switch d.config.MonitorTemperature {
	case "never", "":
		return false
	case "always":
		return true
	case "when_active":
		return d.IsActive() && !d.IsIdle()
	}
	return false
}

// Temperature reads disk temperature (and keeps it in cache for about 1 minute)
func (d *Disk) Temperature() int {
	if d.config.TemperatureSensor == "" {
		return 0
	}
	if !d.TemperatureAvailable() {
		return 0
	}
	if sensor, ok := d.global.TemperatureSensors[d.config.TemperatureSensor]; ok {
		temperature, err := d.temperature.Get(func() (int, error) {
			output, err := sensor.Get(d.expandEnv)
			if err != nil {
				return 0, err
			}
			return output, nil
		})
		if err != nil {
			return 0
		}
		return temperature
	}
	return 0
}

// LastActivity returns the last time the disk has been reading or writing
func (d *Disk) LastActivity() time.Time {
	var err error

	d.activityMutex.Lock()
	defer d.activityMutex.Unlock()

	if d.stats == nil || d.lastActivity.IsZero() {
		d.stats, _ = d.global.GetDiskstats()
		d.lastActivity = time.Now()
		return d.lastActivity
	}
	if d.lastActivity.Add(1 * time.Minute).After(time.Now()) {
		// less than a minute ago, return the value in cache
		return d.lastActivity
	}

	previous := d.stats
	d.stats, err = d.global.GetDiskstats()
	if err != nil {
		d.lastActivity = time.Now()
		return d.lastActivity
	}

	read, write := d.stats.PartitionsIOActivityFrom(previous, filepath.Base(d.Device))
	if read > 0 || write > 0 {
		d.lastActivity = d.stats.timestamp
	}
	return d.lastActivity
}

// IsIdle returns true when the disk hasn't been reading or writing for a set amount of time
func (d *Disk) IsIdle() bool {
	last := d.LastActivity()
	return last.Add(d.idleAfter).Before(time.Now())
}

// HasForceStandby returns true when the disk should be set to standby mode after a period of inactivity
func (d *Disk) HasForceStandby() bool {
	return d.standbyAfter != 0
}

func (d *Disk) StartStandbyWatch() {
	if !d.HasForceStandby() {
		return
	}

	go func() {
		clog.Debugf("will set %s in standby mode after %s of inactivity", d.Device, d.standbyAfter)
		for {
			if d.IsActive() {
				if d.LastActivity().Add(d.standbyAfter).Before(time.Now()) {
					// time to put the disk to sleep
					d.diskStatus.Standby(d.expandEnv)
					d.active.Set(int(enum.DiskStatusActive))
				}
			}
			// default timer is set to 5 minutes
			time.Sleep(5 * time.Minute)
		}
	}()
}

func (d *Disk) expandEnv(input string) string {
	switch input {
	case "DEVICE":
		return d.Device
	case "DEVICE_NAME":
		return filepath.Base(d.Device)
	}
	return "$" + input
}
