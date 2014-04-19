/* Copyright 2014 Ooyala, Inc. All rights reserved.
 *
 * This file is licensed under the MIT license.
 * See the LICENSE file for details.
 */

package htools

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type HConf struct {
	XMLName  xml.Name    `xml:"configuration"`
	PropList []HProperty `xml:"property"`
}

type HProperty struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

func ReadConfFile(minfo MonInfo, filename string) string {
	if _, err := os.Stat(minfo.Confdir); os.IsNotExist(err) {
		log.Printf("configuration file `%s` does not exist.\n", minfo.Confdir)
		os.Exit(1)
	}
	filedata, err2 := ioutil.ReadFile(filepath.Join(minfo.Confdir, filename))
	if err2 != nil {
		panic(err2)
	}
	return string(filedata)
}

func (h HConf) ZooKeepers() string {
	var zookeepers = ""
	for _, prop := range h.PropList {
		if prop.Name == "ha.zookeeper.quorum" {
			zookeepers = prop.Value
		}
	}
	return zookeepers
}

func GetZooKeeperInfo(data string) string {
	var c HConf
	err := xml.Unmarshal([]byte(data), &c)
	if err != nil {
		panic(err)
	}
	return c.ZooKeepers()
}

func GetZooKeepers(minfo MonInfo) string {
	return GetZooKeeperInfo(ReadConfFile(minfo, "core-site.xml"))
}

func (h HConf) ResourceManagers() []*ExtendedResourceManager {
	var resourcemanagers = []*ExtendedResourceManager{}
	var re = regexp.MustCompile("yarn.resourcemanager.address.(.*)")
	var cluster = ""
	var matches = []string{}

	for _, prop := range h.PropList {
		if prop.Name == "yarn.resourcemanager.cluster-id" {
			cluster = prop.Value
		}
	}

	for _, prop := range h.PropList {
		matches = re.FindStringSubmatch(prop.Name)
		if len(matches) > 1 {
			rm := new(ExtendedResourceManager)
			rm.Cluster = cluster
			rm.Id = matches[1]
			rm.Hostname = strings.Split(prop.Value, ":")[0]
			resourcemanagers = append(resourcemanagers, rm)
		}
	}
	return resourcemanagers
}

func GetResourceManagersInfo(data string) []*ExtendedResourceManager {
	var c HConf
	err := xml.Unmarshal([]byte(data), &c)
	if err != nil {
		panic(err)
	}
	return c.ResourceManagers()
}

func GetResourceManagers(minfo MonInfo) []*ExtendedResourceManager {
	return GetResourceManagersInfo(ReadConfFile(minfo, "yarn-site.xml"))
}
