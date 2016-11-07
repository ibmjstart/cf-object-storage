# cf-object-storage

[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> A CloudFoundry Plugin for interacting with OpenStack Object Storage

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

This plugin makes heavy use of the [swiftlygo](https://github.com/ibmjstart/swiftlygo) library. Much more information 
on SLOs and DLOs can be found by reading that library's README.

Additionally, some basic object and container interactions are included as commands. This allows for working with
Object Storage from the command line without having to go through the long authentication process on your own.

## Install

Since this plugin is not currently in an offical Cloud Foundry plugin repo, it will need to be downloaded and installed
manually. 

#### Install From Binary (Recommended)

- Download the binary for your machine ([Linux](https://github.com/ibmjstart/cf-large-objects/tree/master/binaries/linux/cf-large-objects?raw=true), [Mac](https://github.com/ibmjstart/cf-large-objects/tree/master/binaries/darwin/cf-large-objects?raw=true), [Windows](https://github.com/ibmjstart/cf-large-objects/tree/master/binaries/windows/cf-large-objects.exe?raw=true))
- Navigate to the downloaded binary
- Install the plugin with `cf install-plugin cf-large-objects`
- Verify the plugin has been installed with `cf plugins`

**Note:** If you are reinstalling, run `cf uninstall-plugin cf-object-storage` first to uninstall the outdated
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

This plugin is invoked as follows:
`cf os SUBCOMMAND [ARGS...]`

Twelve subcommands are included in this plugin, described below. More information can be found by using `cf os help` followed by any of the subcommands.

#### Subommand List

Subcommand		|Usage															|Description
---		|---															|---
`get-auth-info` | `cf os get-auth-info service_name [-url] [-x]`										|Retrieve a service's x-auth info
`containers` | `cf os containers service_name` | Show all containers in an Object Storage instance
`container-info` | `cf os container-info service_name container_name` | Show a given container's information
`put-container` | `cf os new-container service_name container_name [headers...]` | Create a new container in an Object Storage instance
`rm-container` | `cf os remove-container service_name container_name` | Remove a container from an Object Storage instance
`objects` | `cf os objects service_name container_name` | Show all objects in a container
`object-info` | `cf os object-info service_name container_name object_name` | Show a given object's information
`put-object`    | `cf os put-object service_name container_name path_to_source [-n object_name]` | Upload a file to Object Storage
`get-object` | `cf os get-object service_name container_name object_name` | Download an object from Object Storage
`rm-object` | `cf os rm-object service_name container_name object_name` | Remove an object from a container
`make-dlo`	| `cf os make-dlo service_name dlo_container dlo_name [-c object_container] [-p dlo_prefix]`				|Create a DLO manifest in Object Storage
`make-slo`	| `cf os make-slo service_name slo_container slo_name source_file [-m] [-o output_file] [-s chunk_size] [-t num_threads]`	|Upload a file to Object Storage as an SLO

## Contribute

PRs accepted.

Small note: If editing the README, please conform to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License
Apache 2.0
 © IBM jStart
