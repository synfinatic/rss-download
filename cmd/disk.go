package main

// https://gist.github.com/ttys3/21e2a1215cf1905ab19ddcec03927c75

import (
	"fmt"

	syscall "golang.org/x/sys/unix"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

type DiskStatus struct {
	All   uint64 `json:"all"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
	Avail uint64 `json:"avail"`
}

// disk usage of path/disk
func DiskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Avail = fs.Bavail * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}

func (disk *DiskStatus) DiskInfo() string {
	return fmt.Sprintf("%.2fGB Total, %.2fGB Free",
		float64(disk.Avail)/float64(GB),
		float64(disk.Used)/float64(GB))
}
