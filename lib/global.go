package lib

import (
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/lib/simulation"
)

type Global struct {
	config             cfg.Config
	Disks              map[string]*Disk
	DiskPools          map[string]*DiskPool
	Templates          map[string]*Template
	Tasks              map[string]*Task
	Schedules          map[string]*Schedule
	DiskStatuses       map[string]DiskStatuser
	TemperatureSensors map[string]SensorGetter
	FanControl         *Control
	templ              *template.Template
	diskstats          *Diskstats
	diskstatsMutex     sync.Mutex
}

func NewGlobal(config cfg.Config) (*Global, error) {
	var err error
	global := &Global{
		config:             config,
		DiskPools:          make(map[string]*DiskPool, len(config.DiskPools)),
		Disks:              make(map[string]*Disk, len(config.Disks)),
		Templates:          make(map[string]*Template, len(config.Templates)),
		Tasks:              make(map[string]*Task, len(config.Tasks)),
		TemperatureSensors: make(map[string]SensorGetter, len(config.Sensors)),
		Schedules:          make(map[string]*Schedule, len(config.Schedule)),
		DiskStatuses:       make(map[string]DiskStatuser, len(config.DiskPowerStatus)),
		diskstatsMutex:     sync.Mutex{},
	}

	// Disk pools
	for name, value := range config.DiskPools {
		global.DiskPools[name] = NewDiskPool(global, name, value)
	}

	// Disk Power Status
	for name, value := range config.DiskPowerStatus {
		var diskStatus DiskStatuser
		if config.Simulation {
			diskStatus, err = simulation.NewDiskStatus(name, value)
		} else {
			diskStatus, err = NewDiskStatus(name, value)
		}
		if err != nil {
			return global, err
		}
		global.DiskStatuses[name] = diskStatus
	}

	// Disks
	for name, value := range config.Disks {
		disk, err := NewDisk(global, name, value, global.DiskStatuses)
		if err != nil {
			// display an error but keep going
			clog.Errorf("ignoring configuration for disk %q: %s", name, err)
			continue
		}
		global.Disks[name] = disk
	}

	// Templates
	templates := make([]string, 0, len(config.Templates))
	for name, value := range config.Templates {
		global.Templates[name] = NewTemplate(global, name, value)
		source := value.Source
		if !filepath.IsAbs(source) {
			exec, _ := os.Executable()
			source = filepath.Join(filepath.Dir(exec), source)
		}
		templates = append(templates, source)
	}

	// Now load the templates
	global.templ, err = template.ParseFiles(templates...)
	if err != nil {
		return global, err
	}

	// Temperature sensors
	for sensorName, sensorCfg := range config.Sensors {
		var sensor SensorGetter
		if config.Simulation {
			sensor, err = simulation.NewSensor(sensorName, sensorCfg)
		} else {
			sensor, err = NewSensor(sensorName, sensorCfg, os.DirFS("/"))
		}
		if err != nil {
			return global, err
		}
		global.TemperatureSensors[sensorName] = sensor
	}

	// Tasks
	for name, value := range config.Tasks {
		task, err := NewTask(global, name, value, config.Simulation)
		if err != nil {
			return global, err
		}
		global.Tasks[name] = task
	}

	// Schedules
	for name, value := range config.Schedule {
		global.Schedules[name] = NewSchedule(global, value)
	}

	// Fan controller
	global.FanControl, err = NewControl(global.GetSensorReader, config.FanControl, config.Simulation)
	if err != nil {
		return global, err
	}

	// Disk standby mode
	for _, disk := range global.Disks {
		disk.StartStandbyWatch()
	}

	return global, nil
}

func (g *Global) GetDiskstats() (*Diskstats, error) {
	g.diskstatsMutex.Lock()
	defer g.diskstatsMutex.Unlock()

	if g.diskstats == nil || g.diskstats.timestamp.Add(1*time.Minute).Before(time.Now()) {
		return g.readDiskstats()
	}
	return g.diskstats, nil
}

func (g *Global) readDiskstats() (*Diskstats, error) {
	clog.Debug("reading /proc/diskstats file")

	file, err := os.Open("/proc/diskstats")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	g.diskstats, err = NewDiskstats(file)
	if err != nil {
		return nil, err
	}

	return g.diskstats, nil
}

func (g *Global) GetStartupTasks() []*Task {
	tasks := make([]*Task, 0, len(g.Schedules))
	for _, schedule := range g.Schedules {
		if schedule.OnStartup() {
			tasks = append(tasks, schedule.Task)
		}
	}
	return tasks
}

func (g *Global) GetTimerTasks() []Timer {
	timers := make([]Timer, 0, len(g.Schedules))
	for _, schedule := range g.Schedules {
		duration := schedule.OnTimer()
		if duration > 0 {
			timers = append(timers, Timer{schedule.Task, duration})
		}
	}
	return timers
}

func (g *Global) StartTimers() {
	for _, timer := range g.GetTimerTasks() {
		clog.Debugf("setting up timer task %s every %v", timer.task.Name, timer.every)
		go func(timer Timer) {
			for {
				time.Sleep(timer.every)
				clog.Debugf("running timer task %s", timer.task.Name)
				err := timer.task.Execute()
				if err != nil {
					clog.Error(err)
				}
			}
		}(timer)
	}
}

func (g *Global) GetSensorReader(sensorName string) func() (int, error) {
	// try "standard" sensor first
	if sensor, ok := g.TemperatureSensors[sensorName]; ok {
		return func() (int, error) {
			return sensor.Get(nil)
		}
	}
	// then try disk sensor
	if sensor, ok := g.Disks[sensorName]; ok {
		return func() (int, error) {
			return sensor.Temperature(), nil
		}
	}
	return nil
}
