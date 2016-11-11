#!/bin/sh
cf uninstall-plugin cf-object-storage
rm cf-large-objects
go build
cf install-plugin -f cf-object-storage
