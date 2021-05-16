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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type PushCmd struct {
	Feed    string   `kong:"arg,optional,help='Specify feed name to download'"`
	Filters []string `kong:"arg,optional,help='Specify optional filters to use (default all)'"`
	Cache   string   `kong:"optional,name='cache',short='c',default='${CACHE_FILE}',help='Cache file'"`
}

func (cmd *PushCmd) Run(ctx *RunContext) error {
	allFeeds := ctx.Konf.MapKeys("Feeds")
	feeds := []string{}

	if ctx.Cli.Push.Feed != "" {
		for _, feed := range allFeeds {
			if feed == ctx.Cli.Push.Feed {
				feeds = append(feeds, ctx.Cli.Push.Feed)
				break
			}
		}
		if len(feeds) == 0 {
			return fmt.Errorf("Invalid feed name: %s", ctx.Cli.Push.Feed)
		}
	} else {
		// add our feeds in the specified order
		feedCnt := len(allFeeds)
		for i := 1; i <= feedCnt; i++ {
			for _, feed := range allFeeds {
				order := ctx.Konf.Int(fmt.Sprintf("Feeds.%s.Order", feed))
				if order == i {
					feeds = append(feeds, feed)
				}
			}
		}

		// look for any feeds which don't have an order
		for _, feed := range allFeeds {
			hasOrder := false
			for _, x := range feeds {
				if feed == x {
					hasOrder = true
					break
				}
			}
			if !hasOrder {
				feeds = append(feeds, feed)
			}
		}
	}
	log.Debugf("Feeds = %v", feeds)

	for _, feed := range feeds {
		err := push(ctx, feed)
		if err != nil {
			return err
		}
	}
	return nil
}

func push(ctx *RunContext, feedName string) error {
	log.Infof("Processing: %s", feedName)
	// get our feed
	feedType := ctx.Konf.String(fmt.Sprintf("Feeds.%s.FeedType", feedName))
	if feedType == "" {
		return fmt.Errorf("Missing FeedType for %s", feedName)
	}
	feed, ok := RSS_FEED_TYPES[feedType]
	if !ok {
		return fmt.Errorf("Unknown feed type: %s", feedType)
	}
	feed.Reset()

	err := ctx.Konf.Unmarshal(fmt.Sprintf("Feeds.%s", feedName), feed)
	if err != nil {
		return err
	}
	log.Debugf("Feed: %v", feed)

	// which filters to enable
	filters := []string{}
	if len(ctx.Cli.Push.Filters) != 0 {
		for _, f := range ctx.Cli.Push.Filters {
			filters = append(filters, f)
		}
	} else {
		for filter, _ := range feed.GetFilters() {
			filters = append(filters, filter)
		}
	}

	newEntries, err := DownloadFeed(feedName, RssFeed(feed))
	if err != nil {
		return err
	}

	filteredEntries, err := FilterEntries(newEntries, feed, filters)
	if err != nil {
		return err
	}

	// load our cache
	cacheEntries := []RssFeedEntry{}
	cacheFile := GetPath(ctx.Cli.Push.Cache)
	cacheBytes, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		log.Warnf("Creating new cache file: %s", cacheFile)
	} else {
		json.Unmarshal(cacheBytes, &cacheEntries)
	}

	for _, entry := range filteredEntries {
		if !RssFeedEntryExits(cacheEntries, entry) {
			log.Debugf("New entry: %s", entry.Title)
			var err error = nil
			if feed.GetAutoDownload() {
				err = DownloadUrl(entry, feed)
			} else {
				err = SendPush(ctx.Konf, entry, feed)
			}
			if err != nil {
				log.WithError(err).Errorf("Unable to Download/Push notification for %s", entry.Title)
			} else {
				cacheEntries = append(cacheEntries, entry)
			}
		} else {
			log.Debugf("Entry %s already exists in cache", entry.Title)
		}
	}

	cacheBytes, _ = json.MarshalIndent(cacheEntries, "", "  ")
	return ioutil.WriteFile(cacheFile, cacheBytes, 0644)
}

// Download an entry
func DownloadUrl(entry RssFeedEntry, feed RssFeed) error {
	path := feed.DownloadFilename(feed.GetDownloadPath(), entry)
	log.Debugf("Downloading %s", path)
	resp, err := http.Get(entry.TorrentUrl)
	if err != nil {
		return err
	}
	torrent, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(torrent), 0644)
}
