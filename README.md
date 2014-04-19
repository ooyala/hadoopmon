hadoopmon
=========

High Availability Hadoop (2.x.0) failover notification and hook service.

We needed a way to recieve events (hook) when the failover occured so that we could send alerts on it, as well as handle the failover for web services.

The intention here is regardless of whether you're on a private infrastructure or a public cloud (AWS) that you have a flexible way to handle these type of events.

Based on the zookeeper election, notifies `promote` and `demote` on the current host.

Building
--------

```bash
# Has a dependency on the native zookeeper libraries
$ apt-get install -y zookeeper-native libzookeeper-mt2
$ go get
$ make
```

Usage
-----

```bash
hadoopmon - High Availability Hadoop 2.x.0 failover service

usage:
   hadoopmon [global options] command [command options] [arguments...]

commands:
   namenode, nn   Start the namenode monitor for the given cluster
   resourcemanager, rm  Start the resource-manager monitor for the given cluster
   showbuild, b   Shows the current build information
   help, h    Shows a list of commands or help for one command

global options:
   --conf '/etc/hadoopmon'  The hadoopmon config directory
   --hdir '/etc/hadoop/conf'  The hadoop config directory
   --host 'kiva'		The hostname to assume (override)
   --version      print the version
   --help, -h     show help
```

### Monitoring a namenode

```bash
$ hadoopmon nn titan
```

### Monitoring a resource manager

```bash
$ hadoopmon rm titan
```

#### Options

The `hdir` configuration references the base hadoop configuration directory where `core-site.xml`, `yarn-site.xml`, etc reside.
The `conf` configuration references the base directory for handling `promote` and `demote` commands.
