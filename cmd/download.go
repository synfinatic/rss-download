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

type DownloadCmd struct {
	Feed    string   `kong:"arg,required,help='Specify feed name to download'"`
	Filters []string `kong:"arg,optional,help='Specify optional filter to use (default all)'"`
	Output  string   `kong:"optional,name='output',short='o',default='',help='Output file'"`
	Append  bool     `kong:"optiona,name='append',short='a',default=false,help='Append to existing output file'"`
}

func (cmd *DownloadCmd) Run(ctx *RunContext) error {
	outputFile := fmt.Sprintf("%s.json", ctx.Cli.Download.Feed)
	if ctx.Cli.Download.Output != "" {
		outputFile = ctx.Cli.Download.Output
	}

	// get our feed
	feedType := ctx.Konf.String(fmt.Sprintf("feeds.%s.FeedType", ctx.Cli.Download.Feed))
	feed, ok := RSS_FEED_TYPES[feedType]
	if !ok {
		return fmt.Errorf("Unknown feed type: %s", feedType)
	}
	feed.Reset()

	feedName := fmt.Sprintf("feeds.%s", ctx.Cli.Download.Feed)
	err := ctx.Konf.Unmarshal(feedName, feed)
	if err != nil {
		return err
	}
	log.Debugf("Feed: %v", feed)

	// which filters to enable
	filters := []string{}
	if len(ctx.Cli.Download.Filters) != 0 {
		for _, f := range ctx.Cli.Download.Filters {
			filters = append(filters, f)
		}
	} else {
		for filter, _ := range feed.GetFilters() {
			filters = append(filters, filter)
		}
	}

	newEntries, err := DownloadFeed(ctx.Cli.Download.Feed, feed)
	if err != nil {
		return err
	}

	filteredEntries, err := FilterEntries(newEntries, feed, filters)

	oldEntries := []RssFeedEntry{}
	if ctx.Cli.Download.Append {
		fileBytes, err := ioutil.ReadFile(outputFile)
		if err == nil {
			json.Unmarshal(fileBytes, &oldEntries)
		}
	}

	for _, entry := range filteredEntries {
		if !RssFeedEntryExits(oldEntries, entry) {
			oldEntries = append(oldEntries, entry)
		}
	}
	fileBytes, _ := json.MarshalIndent(oldEntries, "", "  ")
	return ioutil.WriteFile(outputFile, fileBytes, 0644)
}
