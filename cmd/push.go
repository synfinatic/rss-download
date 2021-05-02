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
	"encoding/json"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

type PushCmd struct {
	Feed  string `kong:"arg,required,help='Specify feed name to download'"`
	Cache string `kong:"optional,name='cache',short='c',default='rss-download.json',help='Cache file'"`
}

func (cmd *PushCmd) Run(ctx *RunContext) error {
	// load our cache
	cacheEntries := []RssFeedEntry{}
	cacheBytes, err := ioutil.ReadFile(ctx.Cli.Push.Cache)
	if err != nil {
		log.Warnf("Creating new cache file.")
	} else {
		json.Unmarshal(cacheBytes, &cacheEntries)
	}

	// get our filter
	feedType := ctx.Konf.String(fmt.Sprintf("feeds.%s.FeedType", ctx.Cli.Push.Feed))
	filter, ok := RSS_FEED_TYPES[feedType]
	if !ok {
		return fmt.Errorf("Unknown feed type: %s", feedType)
	}

	feed := fmt.Sprintf("feeds.%s", ctx.Cli.Push.Feed)
	err = ctx.Konf.Unmarshal(feed, filter)
	if err != nil {
		return err
	}
	log.Debugf("Filter: %v", filter)
	newEntries, err := DownloadFeed(ctx.Cli.Push.Feed, RssFeedFilter(filter))
	if err != nil {
		return err
	}

	for _, entry := range newEntries {
		if !RssFeedEntryExits(cacheEntries, entry) {
			log.Debugf("New entry: %s", entry.Title)
			err := SendPush(ctx.Konf, entry, filter)
			if err != nil {
				log.WithError(err).Errorf("Unable to Push notification for %s", entry.Title)
			} else {
				cacheEntries = append(cacheEntries, entry)
			}
		}
	}

	cacheBytes, _ = json.MarshalIndent(cacheEntries, "", "  ")
	return ioutil.WriteFile(ctx.Cli.Push.Cache, cacheBytes, 0644)
}
