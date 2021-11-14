package lib

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// ==  ===================================
//  1  major number
//  2  minor mumber
//  3  device name
//  4  reads completed successfully
//  5  reads merged
//  6  sectors read
//  7  time spent reading (ms)
//  8  writes completed
//  9  writes merged
// 10  sectors written
// 11  time spent writing (ms)
// 12  I/Os currently in progress
// 13  time spent doing I/Os (ms)
// 14  weighted time spent doing I/Os (ms)
// ==  ===================================

// Kernel 4.18+ appends four more fields for discard
// tracking putting the total at 18:

// ==  ===================================
// 15  discards completed successfully
// 16  discards merged
// 17  sectors discarded
// 18  time spent discarding
// ==  ===================================

// Kernel 5.5+ appends two more fields for flush requests:

// ==  =====================================
// 19  flush requests completed successfully
// 20  time spent flushing
// ==  =====================================

const (
	readSuccess           = 3
	readSectors           = 5
	writeSuccess          = 7
	writeSectors          = 9
	inputOutputInProgress = 11
)

type Diskstats struct {
	timestamp time.Time
	data      map[string][]string
}

func NewDiskstats(r io.Reader) (*Diskstats, error) {
	buffer := &bytes.Buffer{}
	_, err := buffer.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(buffer.String(), "\n")
	data := make(map[string][]string, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		line = strings.TrimSpace(line)
		// clean up double spaces
		for strings.Contains(line, "  ") {
			line = strings.ReplaceAll(line, "  ", " ")
		}
		fields := strings.Split(line, " ")
		// minimum number of fields should be 14 (kernel <4.18)
		if len(fields) < 14 {
			return nil, fmt.Errorf("incorrect format: expected 14 or more fields but found %d: %q", len(fields), line)
		}
		data[fields[2]] = fields
	}
	return &Diskstats{
		timestamp: time.Now(),
		data:      data,
	}, nil
}

func (s *Diskstats) IOInProgress(diskname string) int {
	if disk, ok := s.data[diskname]; ok {
		value, err := strconv.Atoi(disk[inputOutputInProgress])
		if err != nil {
			return 0
		}
		return value
	}
	return 0
}

func (s *Diskstats) IOActivityFrom(previous *Diskstats, diskname string) (int64, int64) {
	var read, write int64
	currentRead, currentWrite := s.getSingleReadWriteCounters(diskname)
	previousRead, previousWrite := previous.getSingleReadWriteCounters(diskname)

	if currentRead > 0 {
		read = currentRead - previousRead
	}

	if currentWrite > 0 {
		write = currentWrite - previousWrite
	}

	return read, write
}

func (s *Diskstats) PartitionsIOActivityFrom(previous *Diskstats, diskname string) (int64, int64) {
	var read, write int64
	partitions := s.findPartitions(diskname)
	currentRead, currentWrite := s.getReadWriteCounters(partitions)
	previousRead, previousWrite := previous.getReadWriteCounters(partitions)

	if currentRead > 0 {
		read = currentRead - previousRead
	}

	if currentWrite > 0 {
		write = currentWrite - previousWrite
	}

	return read, write
}

func (s *Diskstats) getReadWriteCounters(partitions []string) (int64, int64) {
	var read, write int64
	for _, partition := range partitions {
		partRead, partWrite := s.getSingleReadWriteCounters(partition)
		read += partRead
		write += partWrite
	}
	return read, write
}

func (s *Diskstats) getSingleReadWriteCounters(diskname string) (int64, int64) {
	var err error
	var read, write int64
	if disk, ok := s.data[diskname]; ok {
		read, err = strconv.ParseInt(disk[readSectors], 10, 64)
		if err != nil {
			return read, write
		}
		write, err = strconv.ParseInt(disk[writeSectors], 10, 64)
		if err != nil {
			return read, write
		}
		return read, write
	}
	return read, write
}

func (s *Diskstats) findPartitions(diskname string) []string {
	partitions := []string{}
	for key := range s.data {
		if key == diskname {
			continue
		}
		if strings.HasPrefix(key, diskname) {
			partitionID := strings.TrimPrefix(key, diskname)
			_, err := strconv.Atoi(partitionID)
			if err != nil {
				continue
			}
			// that's a number behind the diskname => found a partition
			partitions = append(partitions, key)
		}
	}
	return partitions
}
