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
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

const (
	RSS_PARAM_TAG = "param"
)

var RSS_FEED_TYPES = map[string]RssFeed{
	"RFM": &RfmFeed{},
}

// generic RSS entry filter
type RssFilter struct {
	Search     []string `koanf:"Search"`     // regexps
	Categories []string `koanf:"Categories"` // any valid category
	compiled   bool
	match      []*regexp.Regexp
}

// Does the RssFilter have a search regexp which matches the check string?
func (rf *RssFilter) Match(check string) bool {
	if !rf.compiled {
		for i, search := range rf.Search {
			r, err := regexp.Compile(search)
			if err != nil {
				log.WithError(err).Errorf("Unable to compile regexp #%d: %s", i, search)
				rf.match = append(rf.match, nil)
			} else {
				rf.match = append(rf.match, r)
			}
		}
		rf.compiled = true
	}

	for i, match := range rf.match {
		if match == nil {
			continue
		}
		// use compiled
		match := match.Find([]byte(check))
		if match != nil {
			log.Debugf("Matched %s => %s", rf.Search[i], check)
			return true // match
		}
	}
	return false // no match
}

// Does the RssFilter have the given category?
func (rf *RssFilter) HasCategory(category string) bool {
	for _, c := range rf.Categories {
		if c == category {
			return true
		}
	}
	return false
}

// Define the interface for the RSS Feed Filter
type RssFeed interface {
	Reset()
	GetFeedType() string
	GetParam(string) (string, error)
	GenerateUrl() string
	GetPublishFormat() string
	UrlRewriter(string) string
	Match(RssFeedEntry) (bool, string)
	GetFilters() map[string]RssFilter
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
	Description       string    `json:"Description"`
	Url               string    `json:"Url"`
	TorrentUrl        string    `json:"TorrentUrl"`
	TorrentBytes      uint64    `json:"TorrentBytes"`
	TorrentSize       string    `json:"TorrentSize"`
	TorrentCategories []string  `json:"TorrentCategories"`
}

// returns an entry as a pretty string
func (rfe *RssFeedEntry) Sprint() string {
	ret := fmt.Sprintf("Title: %s", rfe.Title)
	ret = fmt.Sprintf("%s\n\tPublished: %s", ret, rfe.Published.Local().Format("2006-01-02 15:04 MST"))
	ret = fmt.Sprintf("%s\n\tCategories: %s", ret, rfe.Categories)
	ret = fmt.Sprintf("%s\n\tDescription: %s", ret, rfe.Description)
	ret = fmt.Sprintf("%s\n\tUrl: %s", ret, rfe.Url)
	ret = fmt.Sprintf("%s\n\tTorrent: %s [%d]", ret, rfe.TorrentUrl, rfe.TorrentBytes)
	ret = fmt.Sprintf("%s\n\tTorrent Categories: %s", ret, strings.Join(rfe.TorrentCategories, ", "))
	ret = fmt.Sprintf("%s\n\tTorrent Size: %s\n", ret, rfe.TorrentSize)
	return ret
}

func DownloadFeed(feedname string, rssFeed RssFeed) ([]RssFeedEntry, error) {
	ret := []RssFeedEntry{}
	url := rssFeed.GenerateUrl()
	log.Debugf("RSS Feed URL = %s", url)
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(url)
	if err != nil {
		return ret, fmt.Errorf("Unable to load %s", url)
	}

	for _, item := range feed.Items {
		t, err := time.Parse(rssFeed.GetPublishFormat(), item.Published)
		if err != nil {
			return ret, fmt.Errorf("Unable to parse Published time `%s` with format `%s`: %s",
				item.Published, rssFeed.GetPublishFormat(), err)
		}

		// figure out torrent info
		torrentUrl := ""
		torrentBytes := uint64(0)

		for _, enclosure := range item.Enclosures {
			if enclosure.Type == "application/x-bittorrent" {
				torrentUrl = enclosure.URL
				torrentBytes, err = strconv.ParseUint(enclosure.Length, 10, 64)
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
				Description:       item.Description,
				Url:               item.Link,
				TorrentUrl:        torrentUrl,
				TorrentBytes:      torrentBytes,
				TorrentSize:       torrentSize,
				TorrentCategories: torrentCategories,
			})
	}
	return ret, nil
}

// filters the given entries and returns those that match our filters
func FilterEntries(entries []RssFeedEntry, feed RssFeed, filters []string) ([]RssFeedEntry, error) {
	retEntries := []RssFeedEntry{}
	for _, entry := range entries {
		// check to see if anything matches
		match, filter := feed.Match(entry)
		if match {
			for _, filterName := range filters {
				// does the hit match one of our specified filters?
				if filter == filterName {
					retEntries = append(retEntries, entry)
				}
			}
		}
	}
	return retEntries, nil
}

// returns true or false if the entry is already in the entries
func RssFeedEntryExits(entries []RssFeedEntry, entry RssFeedEntry) bool {
	for _, e := range entries {
		if e.Title == entry.Title {
			return true
		}
	}
	return false
}
