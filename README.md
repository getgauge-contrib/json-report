json-report
===========
[![Build Status](https://app.snap-ci.com/apoorvam/json-report/branch/master/build_image)](https://app.snap-ci.com/apoorvam/json-report/branch/master)

JSON reporting for [Gauge](http://getgauge.io)

Install through Gauge
---------------------

### Offline installation
* Download the plugin from [Releases](https://github.com/apoorvam/json-report/releases)
```
gauge --install json-report --file <path_to_plugin_zip_file>
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
