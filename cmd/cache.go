package main

/*
 * RSS Download Tool
 * Copyright (c) 2021 Aaron Turner  <synfinatic at gmail dot com>
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
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	ERROR_HOLD_DOWN = 4 // hours
)

type CacheFile struct {
	filename string
	Entries  []RssFeedEntry   `json:"Entries"`
	Errors   map[string]int64 `json:"Errors"`
}

func OpenCache(path string) (*CacheFile, error) {
	cache := CacheFile{
		Entries: []RssFeedEntry{},
		Errors:  map[string]int64{},
	}
	cacheFile := GetPath(path)
	cacheBytes, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		log.Warnf("Creating new cache file: %s", cacheFile)
	} else {
		json.Unmarshal(cacheBytes, &cache)
	}
	cache.filename = cacheFile
	return &cache, nil
}

func (c *CacheFile) SaveCache() error {
	cacheBytes, _ := json.MarshalIndent(*c, "", "  ")
	return ioutil.WriteFile(c.filename, cacheBytes, 0644)
}

// returns true if the error for the given entry is 'new'
func (c *CacheFile) CheckNewError(entry string) bool {
	expire, ok := c.Errors[entry]
	if ok {
		if expire < time.Now().Unix() {
			return true
		}
		return false
	}
	return true
}

func (c *CacheFile) AddError(entry string) {
	c.Errors[entry] = time.Now().Add(time.Hour * ERROR_HOLD_DOWN).Unix()
}
