/* Copyright 2014 Ooyala, Inc. All rights reserved.
 *
 * This file is licensed under the MIT license.
 * See the LICENSE file for details.
 */

package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	. "github.com/ooyala/hadoopmon/htools"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// Note: These are passed in at compile-time.
var (
	VERSION   string
	GITCOMMIT string
)

func main() {
	cli.AppHelpTemplate = `{{.Name}} - {{.Usage}}

usage:
   {{.Name}} [global options] command [command options] [arguments...]

commands:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}
global options:
   {{range .Flags}}{{.}}
   {{end}}
`

	cli.CommandHelpTemplate = `{{.Name}} - {{.Usage}}

usage:
   command {{.Name}} [command options] [arguments...]

description:
   {{.Description}}

options:
   {{range .Flags}}{{.}}
   {{end}}
`

	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Usage = "High Availability Hadoop 2.x.0 failover service"
	app.Version = VERSION

	app.Flags = []cli.Flag{
		cli.StringFlag{"conf", "/etc/hadoopmon", "The hadoopmon config directory"},
		cli.StringFlag{"hdir", "/etc/hadoop/conf", "The hadoop config directory"},
		cli.StringFlag{"host", GetHostname(), "The hostname to assume (override)"},
	}

	app.Commands = []cli.Command{
		{
			Name:      "namenode",
			ShortName: "nn",
			Usage:     "Start the namenode monitor for the given cluster",
			Action: func(c *cli.Context) {
				StartService(c, "namenode")
			},
		},
		{
			Name:      "resourcemanager",
			ShortName: "rm",
			Usage:     "Start the resource-manager monitor for the given cluster",
			Action: func(c *cli.Context) {
				StartService(c, "resource-manager")
			},
		},
		{
			Name:      "showbuild",
			ShortName: "b",
			Usage:     "Shows the current build information",
			Action: func(c *cli.Context) {
				showBuild()
			},
		},
	}

	app.Run(os.Args)
}

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Error determining hostname, assuming default.")
		return "default"
	}
	return hostname
}

func StartService(c *cli.Context, service string) {
	if len(c.Args()) < 1 {
		log.Printf("You must specify a cluster name to monitor")
		os.Exit(1)
	}

	cluster := c.Args()[0]
	confdir := AbsolutePath(c.GlobalString("hdir"))
	mondir := AbsolutePath(c.GlobalString("conf"))
	hostname := c.GlobalString("host")
	minfo := MonInfo{service, cluster, hostname, mondir, confdir}

	go func() {
		StartWatcher(minfo)
	}()

	wait_for_ctrlc()
}

func AbsolutePath(spath string) string {
	pdir, err := filepath.Abs(spath)
	if err != nil {
		log.Printf("error resolving absolute path of `%s`\n", pdir)
		os.Exit(1)
	}
	return pdir
}

func showBuild() {
	fmt.Printf("%s\n", GITCOMMIT)
}

func wait_for_ctrlc() {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-ch
	os.Exit(1)
}
