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
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

const (
	RSS_PARAM_TAG = "param"
)

// Define the interface for the RSS Feed Filter
type RssFeedFilter interface {
	GetFeedType() string
	GetParam(string) (string, error)
	GenerateUrl() string
	GetPublishFormat() string
}

var RSS_FEED_TYPES = map[string]RssFeedFilter{
	"RFM": &RfmFeedFilter{},
}

func GetParamTag(v reflect.Value, fieldName string) (string, error) {
	field, ok := v.Type().FieldByName(fieldName)
	if !ok {
		return "", fmt.Errorf("Invalid field '%s' in %s", fieldName, v.Type().Name())
	}
	tag := string(field.Tag.Get(RSS_PARAM_TAG))
	return tag, nil
}

// Represents a single RSS Feed Entry
type RssFeedEntry struct {
	FeedName          string    `json:"FeedName"`
	Title             string    `json:"Title"`
	Published         time.Time `json:"Published"`
	Categories        []string  `json:"Categories"`
	Url               string    `json:"Url"`
	TorrentUrl        string    `json:"TorrentUrl"`
	TorrentBytes      int64     `json:"TorrentBytes"`
	TorrentSize       string    `json:"TorrentSize"`
	TorrentCategories []string  `json:"TorrentCategories"`
}

// returns an entry as a pretty string
func (rfe *RssFeedEntry) Sprint() string {
	ret := fmt.Sprintf("Title: %s", rfe.Title)
	ret = fmt.Sprintf("%s\n\tPublished: %s", ret, rfe.Published.Local().Format("2006-01-02 15:04 MST"))
	ret = fmt.Sprintf("%s\n\tCategories: %s", ret, rfe.Categories)
	ret = fmt.Sprintf("%s\n\tUrl: %s", ret, rfe.Url)
	ret = fmt.Sprintf("%s\n\tTorrent: %s [%d]", ret, rfe.TorrentUrl, rfe.TorrentBytes)
	ret = fmt.Sprintf("%s\n\tTorrent Categories: %s", ret, strings.Join(rfe.TorrentCategories, ", "))
	ret = fmt.Sprintf("%s\n\tTorrent Size: %s\n", ret, rfe.TorrentSize)
	return ret
}

func DownloadFeed(feedname string, filter RssFeedFilter) ([]RssFeedEntry, error) {
	ret := []RssFeedEntry{}
	url := filter.GenerateUrl()
	log.Errorf("RSS Feed URL = %s", url)
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(url)
	if err != nil {
		return ret, fmt.Errorf("Unable to load %s", url)
	}

	for _, item := range feed.Items {
		t, err := time.Parse(filter.GetPublishFormat(), item.Published)
		if err != nil {
			return ret, fmt.Errorf("Unable to parse Published time `%s` with format `%s`: %s",
				item.Published, filter.GetPublishFormat(), err)
		}

		// figure out torrent info
		torrentUrl := ""
		torrentBytes := int64(0)

		for _, enclosure := range item.Enclosures {
			if enclosure.Type == "application/x-bittorrent" {
				torrentUrl = enclosure.URL
				torrentBytes, err = strconv.ParseInt(enclosure.Length, 10, 64)
				if err != nil {
					return ret, fmt.Errorf("Unable to parse Torrent Bytes `%s`: %s",
						enclosure.Length, err)
				}
				break
			}
		}

		// torrent extension fields
		torrentCategories := []string{}
		torrentSize := ""
		for _, val1 := range item.Extensions {
			for _, val2 := range val1 {
				for _, ext := range val2 {
					switch name := ext.Attrs["name"]; name {
					case "category":
						torrentCategories = strings.Split(ext.Attrs["value"], ", ")
					case "size":
						torrentSize = ext.Attrs["value"]
					}
				}
			}
		}
		ret = append(ret,
			RssFeedEntry{
				FeedName:          feedname,
				Title:             item.Title,
				Published:         t,
				Categories:        item.Categories,
				Url:               item.Link,
				TorrentUrl:        torrentUrl,
				TorrentBytes:      torrentBytes,
				TorrentSize:       torrentSize,
				TorrentCategories: torrentCategories,
			})
	}
	return ret, nil
}
