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
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	RFM_PUBLISH_FORMAT = "2006-01-02 15:04:05"
)

// Impliment the RFM feed filter
type RfmFeed struct {
	FeedType         string
	Order            int                   `koanf:"Order"`
	AutoDownload     bool                  `koanf:"AutoDownload"`
	DownloadPath     string                `koanf:"DownloadPath"`
	BaseUrl          string                `koanf:"BaseUrl"`
	Filters          *map[string]RssFilter `koanf:"Filters"`
	Results          int64                 `koanf:"Results" param:"l"`
	Category         int64                 `koanf:"Category" param:"c"`
	Terms            []string              `koanf:"Terms" param:"s"`
	Uploader         string                `koanf:"Uploader" param:"p"`
	StartDate        string                `koanf:"StartDate" param:"sd"`
	EndDate          string                `koanf:"EndDate" param:"ed"`
	SearchDecription bool                  `koanf:"SearchDecription" param:"d"`
}

// hack around RSS_FEED_TYPES causing stale data to be left around
func (rfm *RfmFeed) Reset() {
	rfm.FeedType = "RFM"
	rfm.AutoDownload = false
	rfm.DownloadPath = ""
	rfm.BaseUrl = ""
	rfm.Order = 0
	rfm.Filters = &map[string]RssFilter{}
	rfm.Results = 0
	rfm.Category = 0
	rfm.Terms = []string{}
	rfm.Uploader = ""
	rfm.StartDate = ""
	rfm.EndDate = ""
	rfm.SearchDecription = false
}

func (rfm *RfmFeed) GetFilters() map[string]RssFilter {
	return *rfm.Filters
}

func (rfm *RfmFeed) DownloadFilename(basePath string, entry RssFeedEntry) string {
	return basePath + fmt.Sprintf("/%s.torrent", entry.Title)
}

func (rfm *RfmFeed) GenerateUrl() string {
	urlParts := []string{}
	if len(rfm.Terms) > 0 {
		p, _ := rfm.GetParam("Terms")
		urlParts = append(urlParts, fmt.Sprintf("%s=%s", p, strings.Join(rfm.Terms, "%20")))
	}
	if rfm.Category != 0 {
		p, _ := rfm.GetParam("Category")
		urlParts = append(urlParts, fmt.Sprintf("%s=%d", p, rfm.Category))
	}
	if rfm.Results != 0 {
		p, _ := rfm.GetParam("Results")
		urlParts = append(urlParts, fmt.Sprintf("%s=%d", p, rfm.Results))
	}
	if rfm.Uploader != "" {
		p, _ := rfm.GetParam("Uploader")
		urlParts = append(urlParts, fmt.Sprintf("%s=%s", p, rfm.Uploader))
	}
	if rfm.StartDate != "" {
		p, _ := rfm.GetParam("StartDate")
		urlParts = append(urlParts, fmt.Sprintf("%s=%s", p, rfm.StartDate))
	}
	if rfm.EndDate != "" {
		p, _ := rfm.GetParam("EndDate")
		urlParts = append(urlParts, fmt.Sprintf("%s=%s", p, rfm.EndDate))
	}
	if rfm.SearchDecription {
		p, _ := rfm.GetParam("SearchDecription")
		urlParts = append(urlParts, "%s=1", p)
	}
	return fmt.Sprintf("%s?%s", rfm.BaseUrl, strings.Join(urlParts, "&"))
}

func (rfm *RfmFeed) GetPublishFormat() string {
	return RFM_PUBLISH_FORMAT
}

func (rfm *RfmFeed) GetFeedType() string {
	return rfm.FeedType
}

func (rfm RfmFeed) GetOrder() int {
	return rfm.Order
}

func (rfm RfmFeed) GetDownloadPath() string {
	return rfm.DownloadPath
}

func (rfm RfmFeed) GetAutoDownload() bool {
	return rfm.AutoDownload
}

func (rfm RfmFeed) GetParam(fieldName string) (string, error) {
	v := reflect.ValueOf(rfm)
	return GetParamTag(v, fieldName)
}

// Need to rewrite the Url field to work on mobile
func (rfm *RfmFeed) UrlRewriter(url string) string {
	re := regexp.MustCompile(`^https://racingfor.me/details/([0-9]+)/.*$`)
	newUrl := re.ReplaceAll([]byte(url), []byte("https://www.racingfor.me//details/${1}"))
	log.Debugf("new url: %s", newUrl)
	return string(newUrl)
}

// Returns if the given entry is a match and if so, which Filter
func (rf *RfmFeed) Match(entry RssFeedEntry) (bool, string) {
	log.Debugf("Looking for match of %s / %s", entry.Title, strings.Join(entry.Categories, ","))
	for fname, filter := range *rf.Filters {
		for _, c := range entry.Categories {
			if filter.HasCategory(c) {
				if filter.Match(entry.Title) || filter.Match(entry.Description) {
					return true, fname
				}
			}
		}
	}
	return false, ""
}
