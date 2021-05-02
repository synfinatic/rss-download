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
	"reflect"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	RFM_PUBLISH_FORMAT = "2006-01-02 15:04:05"
)

// Impliment the RFM feed filter
type RfmFeedFilter struct {
	FeedType         string
	BaseUrl          string   `koanf:"BaseUrl"`
	Terms            []string `koanf:"Terms" param:"s"`
	Category         int64    `koanf:"Category" param:"c"`
	Results          int64    `koanf:"Results" param:"l"`
	Uploader         string   `koanf:"Uploader" param:"p"`
	SearchDecription bool     `koanf:"SearchDecription" param:"d"`
	StartDate        string   `koanf:"StartDate" param:"sd"`
	EndDate          string   `koanf:"EndDate" param:"ed"`
}

func (rfm *RfmFeedFilter) GenerateUrl() string {
	urlParts := []string{}
	if len(rfm.Terms) > 0 {
		p, _ := rfm.GetParam("Terms")
		urlParts = append(urlParts, fmt.Sprintf("%s=%s", p, strings.Join(rfm.Terms, "%20")))
	}
	c, _ := rfm.GetParam("Category")
	r, _ := rfm.GetParam("Results")
	urlParts = append(urlParts, fmt.Sprintf("%s=%d&%s=%d", c, rfm.Category, r, rfm.Results))
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

func (rfm *RfmFeedFilter) GetPublishFormat() string {
	return RFM_PUBLISH_FORMAT
}

func (rfm *RfmFeedFilter) GetFeedType() string {
	return rfm.FeedType
}

func (rfm RfmFeedFilter) GetParam(fieldName string) (string, error) {
	v := reflect.ValueOf(rfm)
	return GetParamTag(v, fieldName)
}

// Need to rewrite the Url field to work on mobile
func (rfm *RfmFeedFilter) UrlRewriter(url string) string {
	re := regexp.MustCompile(`^https://racingfor.me/details/([0-9]+)/.*$`)
	newUrl := re.ReplaceAll([]byte(url), []byte("https://www.racingfor.me//details/${1}"))
	log.Debugf("new url: %s", newUrl)
	return string(newUrl)
}
