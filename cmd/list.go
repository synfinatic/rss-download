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

	log "github.com/sirupsen/logrus"
)

type ListCmd struct {
	Feed string `kong:"arg,optional,help='Specify feed name to list entries'"`
}

func (cmd *ListCmd) Run(ctx *RunContext) error {
	if ctx.Cli.List.Feed == "" {
		return cmd.ListAllFeeds(ctx)
	} else {
		return cmd.ListFeed(ctx)
	}
}

// Just list the feed config
func (cmd *ListCmd) ListAllFeeds(ctx *RunContext) error {
	ctx.Konf.Print()
	return nil
}

// List the contents of the given feed
func (cmd *ListCmd) ListFeed(ctx *RunContext) error {
	feedType := ctx.Konf.String(fmt.Sprintf("Feeds.%s.FeedType", ctx.Cli.List.Feed))
	log.Debugf("FeedType: %s", feedType)

	feed, ok := RSS_FEED_TYPES[feedType]
	if !ok {
		return fmt.Errorf("Unknown feed type: %s", feedType)
	}
	feed.Reset()

	feedPath := fmt.Sprintf("Feeds.%s", ctx.Cli.List.Feed)
	err := ctx.Konf.Unmarshal(feedPath, feed)
	if err != nil {
		return err
	}
	log.Debugf("Feed: %v", feed)

	entries, err := DownloadFeed(ctx.Cli.List.Feed, RssFeed(feed))
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
