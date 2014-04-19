/* Copyright 2014 Ooyala, Inc. All rights reserved.
 *
 * This file is licensed under the MIT license.
 * See the LICENSE file for details.
 */

package htools

import (
	"fmt"
	zk "launchpad.net/gozk/zookeeper"
	"log"
	"os/exec"
	"time"
)

type MonInfo struct {
	Service  string
	Cluster  string
	Hostname string
	Mondir   string
	Confdir  string
}

type ZKResourceManager struct {
	Cluster string
	Id      string
}

type ExtendedResourceManager struct {
	ZKResourceManager
	Hostname string
}

type ZKNameNode struct {
	Cluster  string
	Id       string
	Url      string
	Unknowns []byte
}

const (
	StartMarker        = 10
	ClusterStartMarker = 5
	ClusterEndMarker   = 18
	IdStartMarker      = 3
	IdEndMarker        = 26
	UrlStartMarker     = 20
	UrlEndMarker       = 32
)

func StartWatcher(minfo MonInfo) {
	log.Printf("Watching %s...\n", minfo.Service)
	conn, session, err := zk.Dial(GetZooKeepers(minfo), 5e9)
	if err != nil {
		log.Fatalf("Error connecting to zookeeper: %v", err)
	}

	event := <-session
	if event.State != zk.STATE_CONNECTED {
		log.Fatalf("Can't connect to zookeeper: %v", event)
	}

	switch minfo.Service {
	case "namenode":
		WatchNameNode(conn, minfo)
	case "resource-manager":
		WatchResourceManager(conn, minfo)
	default:
		log.Fatalf("No known service was passed. Exiting...")
	}
}

func GetWatchOn(conn *zk.Conn, minfo MonInfo, node string) (data string, watch <-chan zk.Event) {
	WaitForCreate(conn, minfo, node)
	data, _, watch, err := conn.GetW(node)
	if err != nil {
		fmt.Printf("Error retreiving zookeeper %s info: %v", minfo.Service, err)
	}
	return data, watch
}

func WaitForCreate(conn *zk.Conn, minfo MonInfo, node string) {
	for {
		stat, ewatch, err := conn.ExistsW(node)
		if err != nil {
			fmt.Printf("Error waiting for file zookeeper %s info...", minfo.Service)
			time.Sleep(100 * time.Nanosecond)
		}
		if stat == nil {
			_ = <-ewatch
			break
		} else {
			break
		}
	}
}

func WatchNameNode(conn *zk.Conn, minfo MonInfo) {
	var node = fmt.Sprintf("/hadoop-ha/%s/ActiveStandbyElectorLock", minfo.Cluster)
	minfo.Service = "namenode"

	data, watch := GetWatchOn(conn, minfo, node)
	HandleNameNodeChange(data, minfo)
	go func() {
		for {
			_ = <-watch
			data, watch = GetWatchOn(conn, minfo, node)
			HandleNameNodeChange(data, minfo)
		}
	}()
}

func HandleNameNodeChange(data string, minfo MonInfo) {
	nn := ParseNameNode(data)

	if minfo.Hostname == nn.Url {
		log.Printf("Promoting Host: %s, Leader: %s\n", minfo.Hostname, nn.Url)
		Promote(minfo)
	} else {
		log.Printf("Demoting: Host: %s, Leader: %s\n", minfo.Hostname, nn.Url)
		Demote(minfo)
	}
}

func WatchResourceManager(conn *zk.Conn, minfo MonInfo) {
	var node = fmt.Sprintf("/yarn-leader-election/%s/ActiveStandbyElectorLock", minfo.Cluster)
	minfo.Service = "resource-manager"
	var current_rm = new(ExtendedResourceManager)

	rmlist := GetResourceManagers(minfo)
	for _, rminfo := range rmlist {
		if rminfo.Hostname == minfo.Hostname {
			current_rm = rminfo
			break
		}
	}

	data, watch := GetWatchOn(conn, minfo, node)
	HandleResourceManagerChange(data, current_rm, minfo)
	go func() {
		for {
			_ = <-watch
			data, watch = GetWatchOn(conn, minfo, node)
			HandleResourceManagerChange(data, current_rm, minfo)
		}
	}()
}

func HandleResourceManagerChange(data string, currm *ExtendedResourceManager, minfo MonInfo) {
	rm := ParseResourceManager(data)

	if currm.Id == rm.Id {
		log.Printf("Promoting Host: %v, Leader: %s\n", currm, rm.Id)
		Promote(minfo)
	} else {
		log.Printf("Demoting: Host: %v, Leader: %s\n", currm, rm.Id)
		Demote(minfo)
	}
}

func ParseNameNode(datastring string) (nn *ZKNameNode) {
	var (
		data     = []byte(datastring)
		start    = false
		cstart   = false
		cluster  = []byte{}
		cend     = false
		idstart  = false
		id       = []byte{}
		idend    = false
		ustart   = false
		url      = []byte{}
		urlend   = false
		unknowns = []byte{}
	)

	// Note: Appears hadoop is using thrift or some sort of serialized object.
	for _, element := range data {
		if !start && element == StartMarker {
			start = true
			continue
		}

		if !cstart && element == ClusterStartMarker {
			cstart = true
			continue
		}

		if !cend && element != ClusterEndMarker {
			cluster = append(cluster, element)
			continue
		} else if !cend {
			cend = true
			continue
		}

		if !idstart && element == IdStartMarker {
			idstart = true
			continue
		}

		if !idend && element != IdEndMarker {
			id = append(id, element)
			continue
		} else if !idend {
			idend = true
			continue
		}

		if !ustart && element == UrlStartMarker {
			ustart = true
			continue
		}

		if !urlend && element != UrlEndMarker {
			url = append(url, element)
			continue
		} else if !urlend {
			urlend = true
			continue
		}

		unknowns = append(unknowns, element)
	}

	nn = new(ZKNameNode)
	nn.Cluster = string(cluster)
	nn.Url = string(url)
	nn.Id = string(id)
	return nn
}

func ParseResourceManager(datastring string) (rm *ZKResourceManager) {
	var (
		data    = []byte(datastring)
		start   = false
		cstart  = false
		cluster = []byte{}
		cend    = false
		idstart = false
		id      = []byte{}
	)

	// Note: Appears hadoop is using thrift or some sort of serialized object.
	for _, element := range data {
		if !start && element == StartMarker {
			start = true
			continue
		}

		if !cstart && element == ClusterStartMarker {
			cstart = true
			continue
		}

		if !cend && element != ClusterEndMarker {
			cluster = append(cluster, element)
			continue
		} else if !cend {
			cend = true
			continue
		}

		if !idstart && element == IdStartMarker {
			idstart = true
			continue
		}

		id = append(id, element)
	}

	rm = new(ZKResourceManager)
	rm.Cluster = string(cluster)
	rm.Id = string(id)
	return rm
}

func Promote(minfo MonInfo) {
	cmd := exec.Command(fmt.Sprintf("%s/promote", minfo.Mondir))
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func Demote(minfo MonInfo) {
	cmd := exec.Command(fmt.Sprintf("%s/demote", minfo.Mondir))
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}
