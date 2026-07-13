package lib

import "slices"

type DiskPool struct {
	global *Global
	Name   string
	Disks  []string
}

func NewDiskPool(global *Global, name string, disks []string) *DiskPool {
	for _, disk := range global.Disks {
		if slices.Contains(disks, disk.Name) {
			disk.Pool = name
		}
	}
	return &DiskPool{
		global: global,
		Name:   name,
		Disks:  disks,
	}
}

func (p *DiskPool) CountActive() int {
	count := 0
	for _, diskName := range p.Disks {
		if disk, ok := p.global.Disks[diskName]; ok {
			if disk.IsActive() {
				count++
			}
		}
	}
	return count
}
