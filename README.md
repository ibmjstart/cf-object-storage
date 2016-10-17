# cf-large-objects

[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> A CloudFoundry Plugin for creating Static and Dynamic Large Objects in OpenStack Object Storage

## Table of Contents

- [Background](#background)
- [Install](#install)
- [Usage](#usage)
- [Contribute](#contribute)
- [License](#license)

## Background

Static Large Objects (SLOs) and Dynamic Large Objects (DLOs) are incredibly useful aggregate file types available
in OpenStack Object Storage. However, manipulating them can be quite difficult. This Cloud Foundry CLI plugin is
designed to make using SLOs and DLOs much more accessible. 

This plugin makes heavy use of the [swiftlygo](https://github.com/ibmjstart/swiftlygo) library. Much more information on SLOs and DLOs can be found by reading
that library's README.

## Install

Since this plugin is not currently in an offical Cloud Foundry plugin repo, it will need to be downloaded and installed
manually. 

#### Install From Binary (Recommended)

- Download the binary for your machine ([Linux](https://github.com/ibmjstart/cf-large-objects/tree/master/binaries/linux/cf-large-objects?raw=true), [Mac](https://github.com/ibmjstart/cf-large-objects/tree/master/binaries/darwin/cf-large-objects?raw=true), [Windows](https://github.com/ibmjstart/cf-large-objects/tree/master/binaries/windows/cf-large-objects.exe?raw=true))
- Navigate to the downloaded binary
- Install the plugin with `cf install-plugin cf-large-objects`
- Verify the plugin has been installed with `cf plugins`

**Note:** If you are reinstalling, run `cf uninstall-plugin ObjectStorageLargeObjects` first to uninstall the outdated
version. Additionaly, if installing gives you a permission error run `chmod -x cf-large-objects`.

#### Install From Source

Installing this way requires Go. To download the package, run
```
go get github.com/ibmjstart/cf-large-objects
```

The provided `reinstall.sh` script can then be ran to install the plugin.

**Note:** `reinstall.sh` first attempts to uninstall the plugin, so you may get a failure message from the uninstall
command. This will certainly happen the first time you install. However, as long as the following install succeeds all
should work fine.

## Usage

This plugin provides the user with four new commands, described below. Each pertains to one of the main features of the
swiftlygo library. More information can be found by using `cf help` followed by any of the four commands.

#### Command List

Command		|Usage															|Description
---		|---															|---
`get-auth-info` | `cf get-auth-info service_name [-url] [-x]`										|Retrieve a service's x-auth info
`put-object`    | `cf put-object service_name container_name path_to_source [-n object_name]`						|Upload a file to Object Storage
`make-dlo`	| `cf make-dlo service_name dlo_container dlo_name [-c object_container] [-p dlo_prefix]`				|Create a DLO manifest in Object Storage
`make-slo`	| `cf make-slo service_name slo_container slo_name source_file [-m] [-o output_file] [-s chunk_size] [-t num_threads]`	|Upload a file to Object Storage as an SLO

## Contribute

PRs accepted.

Small note: If editing the README, please conform to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License
Apache 2.0
 © IBM jStart
