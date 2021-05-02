package main

/*
 * RSS Download Manager
 * Copyright (c) 2021 Aaron Turner  <aturner at synfin dot net>
 *
 * This program is free software: you can redistribute it
 * and/or modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or with the authors permission any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import (
	"fmt"
	"strings"
	"time"

	"github.com/gregdel/pushover"
	"github.com/knadh/koanf"
	log "github.com/sirupsen/logrus"
)

const (
	PUSHOVER_USER_KEYS = "Pushover.Users"
	PUSHOVER_DEVICES   = "Pushover.Devices"
	PUSHOVER_APP_KEY   = "Pushover.AppToken"
	PUSHOVER_PRIORITY  = pushover.PriorityNormal
	DISK_PATH          = "DiskPath"
)

func SendPush(konf *koanf.Koanf, entry RssFeedEntry) error {
	appKey := konf.String(PUSHOVER_APP_KEY)
	userKeys := konf.Strings(PUSHOVER_USER_KEYS)
	diskPath := konf.String(DISK_PATH)

	// app and user keys are required
	if appKey == "" {
		return fmt.Errorf("Missing `%s` in config", PUSHOVER_APP_KEY)
	}
	if len(userKeys) == 0 {
		return fmt.Errorf("Missing `%s` in config", PUSHOVER_USER_KEYS)
	}
	diskInfo := ""
	if diskPath != "" {
		disk := DiskUsage(diskPath)
		diskInfo = disk.DiskInfo()
	}

	// if device names are given, use that, otherwise send to all devices
	deviceNamesList := konf.Strings(PUSHOVER_DEVICES)
	deviceNames := ""
	if len(deviceNamesList) > 0 {
		deviceNames = strings.Join(deviceNamesList, ",")
	}

	app := pushover.New(appKey)

	msgText := fmt.Sprintf(`
There is a new %s Torrent available!

Torrent Name: %s

Torrent Size: %s

%s

<a href="%s">More Info</a>
	`, entry.FeedName, entry.Title, entry.TorrentSize, diskInfo, entry.Url)

	msgTitle := entry.Title
	message := pushover.Message{
		Message:     msgText,
		Title:       msgTitle,
		Priority:    PUSHOVER_PRIORITY,
		URL:         entry.TorrentUrl,
		URLTitle:    entry.Title,
		Timestamp:   time.Now().Unix(),
		Retry:       60 * time.Second,
		Expire:      time.Hour,
		DeviceName:  deviceNames,
		CallbackURL: "", // never used
		Sound:       pushover.SoundCosmic,
	}
	for _, user := range userKeys {
		_, err := app.SendMessage(&message, pushover.NewRecipient(user))
		if err != nil {
			log.WithError(err).Errorf("Unable to send message to %s: %s", user, err)
		}
	}
	return nil
}
