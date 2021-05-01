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

	//	"github.com/gregdel/pushover"
	//	log "github.com/sirupsen/logrus"
	"github.com/mmcdole/gofeed"
)

type ListCmd struct {
	Feed string `kong:"arg,optional,help='Specify feed name to list entries'"`
}

func (cmd *ListCmd) Run(ctx *RunContext) error {
	if ctx.Cli.List.Feed == "" {
		return cmd.ListAllFeeds(ctx)
	} else {
		return cmd.ListFeed(ctx, ctx.Cli.List.Feed)
	}
}

// Just list the feed config
func (cmd *ListCmd) ListAllFeeds(ctx *RunContext) error {
	feeds := ctx.Konf.MapKeys("feeds")
	total := len(feeds)
	for i, feed := range feeds {
		fmt.Printf("%s:\n", feed)
		thisFeed := fmt.Sprintf("feeds.%s", feed)
		for k, v := range ctx.Konf.StringMap(thisFeed) {
			fmt.Printf("\t%s: %s\n", k, v)
		}
		if i+1 < total {
			fmt.Printf("\n")
		}
	}

	return nil

}

// List the contents of the given feed
func (cmd *ListCmd) ListFeed(ctx *RunContext, feed string) error {
	fp := gofeed.NewParser()
	selector := fmt.Sprintf("feeds.%s.url", feed)
	url := ctx.Konf.String(selector)
	urlFeed, err := fp.ParseURL(url)
	if err != nil {
		return fmt.Errorf("Unable to load %s: %s", url, feed)
	}

	for i, item := range urlFeed.Items {
		fmt.Printf("%d\tTitle: %s\n", i, item.Title)
		fmt.Printf("\tPubDate: %s\n", item.Published)
		// also item.PublishedParsed => *time.Time
		fmt.Printf("\tCategory: %s\n", strings.Join(item.Categories, ", "))
		fmt.Printf("\tUrl: %s\n", item.Link)
		for _, enclosure := range item.Enclosures {
			if enclosure.Type == "application/x-bittorrent" {
				fmt.Printf("\tTorrent: %v [%s]\n", enclosure.URL, enclosure.Length)
			}
		}
		for key, val := range item.Custom {
			fmt.Printf("\tCustom: %s => %s", key, val)
		}
		for x, val1 := range item.Extensions {
			for y, val2 := range val1 {
				for _, ext := range val2 {
					fmt.Printf("\t\t%s:%s %s => %s\n", x, y, ext.Attrs["name"], ext.Attrs["value"])
				}
			}
		}
		fmt.Printf("\n")
	}

	return nil
}
