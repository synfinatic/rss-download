package main

/*
 * RSS Download Tool
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

func SendPush(konf *koanf.Koanf, entry RssFeedEntry, feed RssFeed) error {
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
		disk, err := DiskUsage(konf, diskPath)
		if err != nil {
			return err
		}
		diskInfo = disk.DiskInfo(entry.TorrentBytes)
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

<a href="https://www.synfin.net/transmission/web/">Highlandpark Transmission</a>

<a href="https://brix.int.synfin.net/transmission/web/">Brix Transmission</a>
	`, entry.FeedName, entry.Title, entry.TorrentSize, diskInfo, feed.UrlRewriter(entry.Url))

	msgTitle := entry.Title
	message := pushover.Message{
		HTML:        true,
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

func SendPushError(konf *koanf.Koanf, err error) error {
	appKey := konf.String(PUSHOVER_APP_KEY)
	userKeys := konf.Strings(PUSHOVER_USER_KEYS)

	msgText := fmt.Sprintf(`
Torrent Error:

%s
	`, err)

	// app and user keys are required
	if appKey == "" {
		return fmt.Errorf("Missing `%s` in config", PUSHOVER_APP_KEY)
	}
	if len(userKeys) == 0 {
		return fmt.Errorf("Missing `%s` in config", PUSHOVER_USER_KEYS)
	}

	// if device names are given, use that, otherwise send to all devices
	deviceNamesList := konf.Strings(PUSHOVER_DEVICES)
	deviceNames := ""
	if len(deviceNamesList) > 0 {
		deviceNames = strings.Join(deviceNamesList, ",")
	}

	app := pushover.New(appKey)
	message := pushover.Message{
		HTML:        false,
		Message:     msgText,
		Title:       "RSS Feed Error",
		Priority:    PUSHOVER_PRIORITY,
		Timestamp:   time.Now().Unix(),
		Retry:       60 * time.Second,
		Expire:      time.Hour,
		DeviceName:  deviceNames,
		CallbackURL: "",
		Sound:       pushover.SoundSpaceAlarm,
	}

	for _, user := range userKeys {
		_, err := app.SendMessage(&message, pushover.NewRecipient(user))
		if err != nil {
			log.WithError(err).Errorf("Unable to send message to %s: %s", user, err)
		}
	}
	return nil
}
