package lib

type DiskPool struct {
	global *Global
	Name   string
	Disks  []string
}

func NewDiskPool(global *Global, name string, disks []string) *DiskPool {
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
