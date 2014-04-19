/* Copyright 2014 Ooyala, Inc. All rights reserved.
 *
 * This file is licensed under the MIT license.
 * See the LICENSE file for details.
 */

package htools

import (
	. "launchpad.net/gocheck"
	"testing"
)

func TestConfig(t *testing.T) { TestingT(t) }

type ParserSuite struct{}

var _ = Suite(&ParserSuite{})

func (s *ParserSuite) TestGetZooKeepers(c *C) {
	data := `
<configuration>
  <property>
    <name>dfs.namenode.secondary.http-address</name>
    <value>cdh5-namenode2.mycompany.com:50070</value>
  </property>

  <property>
    <name>dfs.datanode.dns.nameserver</name>
    <value>10.10.0.1</value>
  </property>

  <property>
    <name>ha.zookeeper.quorum</name>
    <value>10.20.10.1:2181,10.20.11.2:2181,10.20.12.1:2181,10.20.13.1:2181,10.20.14.1:2181,10.20.15.1:2181</value>
  </property>
</configuration>
`
	zkstring := "10.20.10.1:2181,10.20.11.2:2181,10.20.12.1:2181,10.20.13.1:2181,10.20.14.1:2181,10.20.15.1:2181"
	c.Assert(GetZooKeeperInfo(data), Equals, zkstring)
}

func (s *ParserSuite) TestGetResourceManagers(c *C) {
	data := `
<configuration>
  <property>
    <name>yarn.resourcemanager.cluster-id</name>
    <value>titan</value>
  </property>

	  <!-- rm1 configuration -->
  <property>
    <name>yarn.resourcemanager.address.rm1</name>
    <value>myfirstserver.mycompany.com:8032</value>
  </property>

  <property>
    <name>yarn.resourcemanager.scheduler.address.rm1</name>
    <value>myfirstserver.mycompany.com:8030</value>
  </property>

  <property>
    <name>yarn.resourcemanager.webapp.address.rm1</name>
    <value>myfirstserver.mycompany.com:8088</value>
  </property>

  <property>
    <name>yarn.resourcemanager.resource-tracker.address.rm1</name>
    <value>myfirstserver.mycompany.com:8031</value>
  </property>

  <property>
    <name>yarn.resourcemanager.admin.address.rm1</name>
    <value>myfirstserver.mycompany.com:8033</value>
  </property>

  <property>
    <name>yarn.resourcemanager.ha.admin.address.rm1</name>
    <value>myfirstserver.mycompany.com:8034</value>
  </property>

  <!-- rm2 configuration -->
  <property>
    <name>yarn.resourcemanager.address.rm2</name>
    <value>mysecondserver.mycompany.com:8032</value>
  </property>

  <property>
    <name>yarn.resourcemanager.scheduler.address.rm2</name>
    <value>mysecondserver.mycompany.com:8030</value>
  </property>

  <property>
    <name>yarn.resourcemanager.webapp.address.rm2</name>
    <value>mysecondserver.mycompany.com:8088</value>
  </property>

  <property>
    <name>yarn.resourcemanager.resource-tracker.address.rm2</name>
    <value>mysecondserver.mycompany.com:8031</value>
  </property>

  <property>
    <name>yarn.resourcemanager.admin.address.rm2</name>
    <value>mysecondserver.mycompany.com:8033</value>
  </property>

  <property>
    <name>yarn.resourcemanager.ha.admin.address.rm2</name>
    <value>mysecondserver.mycompany.com:8034</value>
  </property>

  <!-- log configuration -->
</configuration>
`

	rm1 := new(ExtendedResourceManager)
	rm1.Id = "rm1"
	rm1.Cluster = "titan"
	rm1.Hostname = "myfirstserver.mycompany.com"

	rm2 := new(ExtendedResourceManager)
	rm2.Id = "rm2"
	rm2.Cluster = "titan"
	rm2.Hostname = "mysecondserver.mycompany.com"

	rmlist := []*ExtendedResourceManager{rm1, rm2}
	rminfo := GetResourceManagersInfo(data)

	c.Assert(len(rminfo), Equals, 2)

	c.Assert(rminfo[0].Id, Equals, rmlist[0].Id)
	c.Assert(rminfo[0].Cluster, Equals, rmlist[0].Cluster)
	c.Assert(rminfo[0].Hostname, Equals, rmlist[0].Hostname)

	c.Assert(rminfo[1].Id, Equals, rmlist[1].Id)
	c.Assert(rminfo[1].Cluster, Equals, rmlist[1].Cluster)
	c.Assert(rminfo[1].Hostname, Equals, rmlist[1].Hostname)
}

func (s *ParserSuite) TestParseNameNode(c *C) {
	// Example Hadoop Serialized Namenode object
	nndata := []byte{
		10, 5, 116, 105, 116, 97, 110, 18, 3, 110, 110, 49, 26, 20,
		115, 111, 109, 101, 114, 97, 110, 100, 111, 109, 117, 114,
		108, 46, 99, 111, 109, 32, 212, 62, 40, 211, 62}

	nn := ParseNameNode(string(nndata))
	c.Assert(nn.Cluster, Equals, "titan")
	c.Assert(nn.Id, Equals, "nn1")
	c.Assert(nn.Url, Equals, "somerandomurl.com")
}

func (s *ParserSuite) TestParseResourceManager(c *C) {
	// Example Hadoop Serialized ResourceManager object
	rmdata := []byte{10, 5, 116, 105, 116, 97, 110, 18, 3, 114, 109, 50}

	rm := ParseResourceManager(string(rmdata))
	c.Assert(rm.Cluster, Equals, "titan")
	c.Assert(rm.Id, Equals, "rm2")
}
