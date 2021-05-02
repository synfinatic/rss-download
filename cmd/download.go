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
	Feed   string `kong:"arg,required,help='Specify feed name to download'"`
	Output string `kong:"optional,name='output',short='o',default='',help='Output file'"`
	Append bool   `kong:"optiona,name='append',short='a',default=false,help='Append to existing output file'"`
}

func (cmd *DownloadCmd) Run(ctx *RunContext) error {
	feed := fmt.Sprintf("feeds.%s", ctx.Cli.Download.Feed)
	feedType := ctx.Konf.String(fmt.Sprintf("feeds.%s.FeedType", ctx.Cli.Download.Feed))
	log.Debugf("FeedType: %s", feedType)

	outputFile := fmt.Sprintf("%s.json", ctx.Cli.Download.Feed)
	if ctx.Cli.Download.Output != "" {
		outputFile = ctx.Cli.Download.Output
	}

	filter, ok := RSS_FEED_TYPES[feedType]
	if !ok {
		return fmt.Errorf("Unknown feed type: %s", feedType)
	}

	err := ctx.Konf.Unmarshal(feed, filter)
	if err != nil {
		return err
	}
	log.Debugf("Filter: %v", filter)

	entries, err := DownloadFeed(ctx.Cli.Download.Feed, RssFeedFilter(filter))
	if err != nil {
		return err
	}

	oldEntries := []RssFeedEntry{}
	if ctx.Cli.Download.Append {
		fileBytes, err := ioutil.ReadFile(outputFile)
		if err == nil {
			json.Unmarshal(fileBytes, &oldEntries)
		}
	}

	for _, entry := range entries {
		if !RssFeedEntryExits(oldEntries, entry) {
			oldEntries = append(oldEntries, entry)
		}
	}
	fileBytes, _ := json.MarshalIndent(oldEntries, "", "  ")
	return ioutil.WriteFile(outputFile, fileBytes, 0644)
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
