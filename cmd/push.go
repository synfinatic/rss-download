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
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

type PushCmd struct {
	Feed    string   `kong:"arg,optional,help='Specify feed name to download'"`
	Filters []string `kong:"arg,optional,help='Specify optional filters to use (default all)'"`
	Cache   string   `kong:"optional,name='cache',short='c',default='rss-download.json',help='Cache file'"`
}

func (cmd *PushCmd) Run(ctx *RunContext) error {
	allFeeds := ctx.Konf.MapKeys("feeds")
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
				order := ctx.Konf.Int(fmt.Sprintf("feeds.%s.Order", feed))
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
	log.Debugf("feeds = %v", feeds)

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
	feedType := ctx.Konf.String(fmt.Sprintf("feeds.%s.FeedType", feedName))
	if feedType == "" {
		return fmt.Errorf("Missing FeedType for %s", feedName)
	}
	feed, ok := RSS_FEED_TYPES[feedType]
	if !ok {
		return fmt.Errorf("Unknown feed type: %s", feedType)
	}
	feed.Reset()

	err := ctx.Konf.Unmarshal(fmt.Sprintf("feeds.%s", feedName), feed)
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
	cacheBytes, err := ioutil.ReadFile(ctx.Cli.Push.Cache)
	if err != nil {
		log.Warnf("Creating new cache file.")
	} else {
		json.Unmarshal(cacheBytes, &cacheEntries)
	}

	for _, entry := range filteredEntries {
		if !RssFeedEntryExits(cacheEntries, entry) {
			log.Debugf("New entry: %s", entry.Title)
			err := SendPush(ctx.Konf, entry, feed)
			if err != nil {
				log.WithError(err).Errorf("Unable to Push notification for %s", entry.Title)
			} else {
				cacheEntries = append(cacheEntries, entry)
			}
		} else {
			log.Debugf("Entry %s already exists in cache", entry.Title)
		}
	}

	cacheBytes, _ = json.MarshalIndent(cacheEntries, "", "  ")
	return ioutil.WriteFile(ctx.Cli.Push.Cache, cacheBytes, 0644)
}
