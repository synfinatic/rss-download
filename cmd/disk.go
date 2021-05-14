package main

// https://gist.github.com/ttys3/21e2a1215cf1905ab19ddcec03927c75

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
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
		log.WithError(err).Errorf("Unable to Statfs: %s", path)
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Avail = fs.Bavail * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	info, err := os.Lstat(path)
	if err != nil {
		log.WithError(err).Errorf("Unable to Lstat: %s", path)
		return
	}
	disk.Used, err = GetDirectorySize(path, info)
	if err != nil {
		log.WithError(err).Errorf("Unable to get directory size: '%s'", path)
	}
	return
}

func GetDirectorySize(path string, info os.FileInfo) (uint64, error) {
	var size uint64

	dir, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return 0, err
	}

	for _, file := range files {
		if file.Name() == "lost+found" {
			continue
		}
		if file.IsDir() {
			subdir := fmt.Sprintf("%s/%s", path, file.Name())
			s, err := GetDirectorySize(subdir, file)
			if err != nil {
				return 0, fmt.Errorf("failed recurse %s: %s", subdir, err)
			}
			size += s
		} else {
			size += uint64(file.Size())
		}
	}

	return size, nil
}

func (disk *DiskStatus) DiskInfo(newFileSize uint64) string {
	color := "#ff0000" // red
	if disk.Avail > (newFileSize + uint64(5*MB)) {
		color = "#00ff00" // green
	}
	return fmt.Sprintf(`<font color="%s">%.2fGB Free, %.2fGB Used</font>`,
		color,
		float64(disk.Avail)/float64(GB),
		float64(disk.Used)/float64(GB))
}
