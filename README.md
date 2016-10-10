cf-large-objects: a Cloud Foundry CLI plugin
=====================
A plugin for interacting with objects in Object Storage.

## Command List

Command		|Usage															|Description
---		|---															|---
`get-auth-info` | `cf get-auth-info service_name [-url] [-x]`										|Retrieve a service's x-auth info
`put-object`    | `cf put-object service_name container_name path_to_source [-n object_name]`						|Upload a file to Object Storage
`make-dlo`	| `cf make-dlo service_name dlo_container dlo_name [-c object_container] [-p dlo_prefix]`				|Create a DLO manifest in Object Storage
`make-slo`	| `cf make-slo service_name slo_container slo_name source_file [-m] [-o output_file] [-s chunk_size] [-t num_threads]`	|Upload a file to Object Storage as an SLO
