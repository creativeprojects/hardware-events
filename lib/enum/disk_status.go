package enum

type DiskStatus int

const (
	DiskStatusUnknown DiskStatus = iota
	DiskStatusActive
	DiskStatusStandby
	DiskStatusSleeping
)
