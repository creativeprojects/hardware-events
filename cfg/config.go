package cfg

import (
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config from the file
type Config struct {
	Simulation      bool                `yaml:"simulation"`
	DiskPowerStatus DiskPowerStatus     `yaml:"disk_power_status"`
	Sensors         map[string]Task     `yaml:"sensors"`
	DiskPools       map[string][]string `yaml:"disk_pools"`
	Disks           map[string]Disk     `yaml:"disks"`
	Templates       map[string]Template `yaml:"templates"`
	Tasks           map[string]Task     `yaml:"tasks"`
	Schedule        map[string]Schedule `yaml:"schedule"`
	FanControl      FanControl          `yaml:"fan_control"`
}

type DiskPowerStatus struct {
	CheckCommand   string `yaml:"check_command"`
	Active         string `yaml:"active"`
	Standby        string `yaml:"standby"`
	Sleeping       string `yaml:"sleeping"`
	StandbyCommand string `yaml:"standby_command"`
	Timeout        string `yaml:"timeout"`
}

// Disk configuration
type Disk struct {
	Device             string `yaml:"device"`
	TemperatureSensor  string `yaml:"temperature_sensor"`
	MonitorTemperature string `yaml:"monitor_temperature"`
	LastActive         string `yaml:"last_active"`
	StandbyAfter       string `yaml:"standby_after"`
}

type Template struct {
	Source string `yaml:"source"`
}

type Task struct {
	Command     string   `yaml:"command"`
	File        string   `yaml:"file"`
	Files       []string `yaml:"files"`
	Stdin       Source   `yaml:"stdin"`
	Regexp      string   `yaml:"regexp"`
	Divider     int      `yaml:"divider"`
	Aggregation string   `yaml:"aggregation"`
	Timeout     string   `yaml:"timeout"`
}

type Source struct {
	Template string `yaml:"template"`
}

type Schedule struct {
	Task string   `yaml:"task"`
	When []string `yaml:"when"`
}

type FanControl struct {
	InitCommand string               `yaml:"init_command"`
	SetCommand  string               `yaml:"set_command"`
	ExitCommand string               `yaml:"exit_command"`
	Timeout     string               `yaml:"timeout"`
	Parameters  map[string]Parameter `yaml:"parameters"`
	Zones       map[string]FanZone   `yaml:"zones"`
}

type Parameter struct {
	Format string `yaml:"format"`
}

type FanZone struct {
	ID           int               `yaml:"id"`
	MinSpeed     int               `yaml:"min_speed"`
	MaxSpeed     int               `yaml:"max_speed"`
	DefaultSpeed int               `yaml:"default_speed"`
	RunEvery     string            `yaml:"run_every"`
	Sensors      map[string]Sensor `yaml:"sensors"`
}

type Sensor struct {
	Average  string       `yaml:"average"`
	RunEvery string       `yaml:"run_every"`
	Rules    []SensorRule `yaml:"rules"`
}

type SensorRule struct {
	Temperature FromTo    `yaml:"temperature"`
	Fan         SetFromTo `yaml:"fan_speed"`
	RunEvery    string    `yaml:"run_every"`
}

type FromTo struct {
	From int `yaml:"from"`
	To   int `yaml:"to"`
}

type SetFromTo struct {
	Set  int `yaml:"set"`
	From int `yaml:"from"`
	To   int `yaml:"to"`
}

// LoadFileConfig loads the configuration from the file
func LoadFileConfig(fileName string) (Config, error) {
	if !filepath.IsAbs(fileName) {
		exec, err := os.Executable()
		if err != nil {
			return Config{}, err
		}
		fileName = filepath.Join(filepath.Dir(exec), fileName)
	}
	file, err := os.Open(fileName)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()
	return loadConfig(file)
}

func loadConfig(reader io.Reader) (Config, error) {
	config := Config{}
	decoder := yaml.NewDecoder(reader)
	err := decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}
