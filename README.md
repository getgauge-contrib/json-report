json-report
===========
[![Build Status](https://travis-ci.org/getgauge-contrib/json-report.svg?branch=master)](https://travis-ci.org/getgauge-contrib/json-report)

JSON reporting plugin for [Gauge](http://gauge.org)

Installation
------------

### Install through Gauge 

```
gauge install json-report
```
Installing specific version:

```
gauge install json-report --version 0.1.0
```

### Offline installation
* Download the plugin from [Releases](https://github.com/getgauge-contrib/json-report/releases)
```
gauge install json-report --file <path_to_plugin_zip_file>
```

### Usage

Add this plugin to your Gauge project by registering it in `manifest.json` file. You can also do this by:

```
gauge install json-report
```

By default, reports are generated in `reports/json-report` directory of your Gauge project. You can set a custom location by setting the below mentioned property in `default.properties` file of `env/default` directory.

```
#The path to the gauge reports directory. Should be either relative to the project directory or an absolute path
gauge_reports_dir = reports
```

You can also choose to override the reports after each execution or retain all of them as follows.

```
#Set as false if gauge reports should not be overwritten on each execution. A new time-stamped directory will be created on each execution.
overwrite_reports = true
```

Build from Source
-----------------

### Requirements
* [Golang](http://golang.org/)

### Compiling

```
go run build/make.go
```

For cross-platform compilation

```
go run build/make.go --all-platforms
```

### Installing
After compilation

```
go run build/make.go --install
```

Installing to a CUSTOM_LOCATION

```
go run build/make.go --install --plugin-prefix CUSTOM_LOCATION
```

### Creating distributable

Note: Run after compiling

```
go run build/make.go --distro
```

For distributable across platforms: Windows and Linux for both x86 and x86_64

```
go run build/make.go --distro --all-platforms
```
