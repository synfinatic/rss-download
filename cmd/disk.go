package main

// https://gist.github.com/ttys3/21e2a1215cf1905ab19ddcec03927c75

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/knadh/koanf"
	log "github.com/sirupsen/logrus"
	syscall "golang.org/x/sys/unix"
)

const (
	B           = 1
	KB          = 1024 * B
	MB          = 1024 * KB
	GB          = 1024 * MB
	TB          = 1024 * GB
	DISK_BUFFER = "DiskBuffer"
)

type DiskStatus struct {
	All   uint64 `json:"all"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
	Avail uint64 `json:"avail"`
}

// disk usage of path/disk
func DiskUsage(konf *koanf.Koanf, path string) (DiskStatus, error) {
	diskBuffer, err := convertBytesString(konf.String(DISK_BUFFER))
	if err != nil {
		log.WithError(err).Errorf("Unable to apply %s", DISK_BUFFER)
		diskBuffer = 0
	}

	disk := DiskStatus{}

	fs := syscall.Statfs_t{}
	err = syscall.Statfs(path, &fs)
	if err != nil {
		return DiskStatus{}, fmt.Errorf("Unable to statfs: %s: %s", path, err)
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Avail = fs.Bavail * uint64(fs.Bsize)

	// apply our buffer
	if disk.Avail < diskBuffer {
		disk.Avail = 0
	} else {
		disk.Avail -= diskBuffer
	}

	disk.Free = fs.Bfree * uint64(fs.Bsize)

	info, err := os.Lstat(path)
	if err != nil {
		return DiskStatus{}, fmt.Errorf("Unable to Lstat %s: %s", path, err.Error())
	}
	disk.Used, err = GetDirectorySize(path, info)
	if err != nil {
		return DiskStatus{}, fmt.Errorf("Unable to get directory size %s: %s", path, err.Error())
	}
	return disk, nil
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

// Converts XXGB/TB/MB/KB to number of bytes
func convertBytesString(str string) (uint64, error) {
	if str == "" {
		return 0, nil
	}
	var zBytes uint64
	var err error

	if strings.HasSuffix(str, "TB") {
		tb := strings.Trim(str, "TB")
		zBytes, err = strconv.ParseUint(tb, 10, 64)
		zBytes *= TB
	} else if strings.HasSuffix(str, "GB") {
		gb := strings.Trim(str, "GB")
		zBytes, err = strconv.ParseUint(gb, 10, 64)
		zBytes *= GB
	} else if strings.HasSuffix(str, "MB") {
		mb := strings.Trim(str, "MB")
		zBytes, err = strconv.ParseUint(mb, 10, 64)
		zBytes *= MB
	} else if strings.HasSuffix(str, "KB") {
		kb := strings.Trim(str, "KB")
		zBytes, err = strconv.ParseUint(kb, 10, 64)
		zBytes *= KB
	} else if strings.HasSuffix(str, "B") {
		return 0, fmt.Errorf("Unparsable bytes string: %s", str)
	}

	return zBytes, err
}
