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
	"os"

	"github.com/alecthomas/kong"
	//	"github.com/gregdel/pushover"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
)

var Version = "unknown"
var Buildinfos = "unknown"
var Tag = "NO-TAG"
var CommitID = "unknown"

type RunContext struct {
	Ctx  *kong.Context
	Cli  *CLI
	Konf *koanf.Koanf
}

type CLI struct {
	// Common Arguments
	LogLevel string `kong:"optional,short='L',name='loglevel',default='info',enum='error,warn,info,debug',help='Logging level [error|warn|info|debug]'"`
	Lines    bool   `kong:"optional,name='lines',default=false,help='Include line numbers in logs'"`
	Log      string `kong:"optional,name='log',default='stderr',help='Output log file'"`
	Config   string `kong:"required,name='config',default='rssdownload.yaml',help='Config file'"`

	// sub commands
	Version  VersionCmd  `kong:"cmd,help='Print version and exit'"`
	Download DownloadCmd `kong:"cmd,help='Download the feeds'"`
	List     ListCmd     `kong:"cmd,help='List the configured feeds'"`
}

func main() {
	k := kong.Description("RSS Download Manager")
	cli := CLI{}
	ctx := kong.Parse(&cli, k)

	switch cli.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
		log.SetOutput(colorable.NewColorableStdout())
	case "warn":
		log.SetLevel(log.WarnLevel)
		log.SetOutput(colorable.NewColorableStdout())
	case "error":
		log.SetLevel(log.ErrorLevel)
		log.SetOutput(colorable.NewColorableStdout())
	}
	if cli.Lines {
		log.SetReportCaller(true)
	}
	if cli.Log == "stderr" {
		log.SetOutput(os.Stderr)
	} else {
		file, err := os.OpenFile(cli.Log, os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			log.WithError(err).Fatalf("Unable to open log file: %s", cli.Log)
		}
		log.SetOutput(file)
	}

	run_ctx := RunContext{
		Ctx:  ctx,
		Cli:  &cli,
		Konf: koanf.New("."),
	}

	if err := run_ctx.Konf.Load(file.Provider(cli.Config), yaml.Parser()); err != nil {
		log.WithError(err).Fatalf("Unable to open config file")
	}

	err := ctx.Run(&run_ctx)
	if err != nil {
		log.Panicf("Error running command: %s", err.Error())
	}

}

// Version Command
type VersionCmd struct{}

func (cmd *VersionCmd) Run(ctx *RunContext) error {
	fmt.Printf("RSS Download Manager v%s -- Copyright 2021 Aaron Turner\n", Version)
	fmt.Printf("%s (%s) built at %s\n", CommitID, Tag, Buildinfos)
	return nil
}
