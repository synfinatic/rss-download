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
	log "github.com/sirupsen/logrus"
)

type DownloadCmd struct {
	Feed string `kong:"arg,required,help='Specify feed name to download'"`
}

func (cmd *DownloadCmd) Run(ctx *RunContext) error {
	feedType := ctx.Konf.String(fmt.Sprintf("feeds.%s.FeedType", ctx.Cli.Download.Feed))
	log.Debugf("FeedType: %s", feedType)

	filter, ok := RSS_FEED_TYPES[feedType]
	if !ok {
		return fmt.Errorf("Unknown feed type: %s", feedType)
	}

	feed := fmt.Sprintf("feeds.%s", ctx.Cli.Download.Feed)
	err := ctx.Konf.Unmarshal(feed, filter)
	if err != nil {
		return err
	}
	log.Debugf("Filter: %v", filter)

	entries, err := DownloadFeed(RssFeedFilter(filter))
	if err != nil {
		return err
	}

	total := len(entries)
	for i, entry := range entries {
		fmt.Printf("%d %s", i, entry.Sprint())
		if i+1 < total {
			fmt.Printf("\n")
		}
	}
	return nil
}
