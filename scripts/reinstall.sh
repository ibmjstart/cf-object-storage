#!/bin/sh
cf uninstall-plugin cf-object-storage
rm cf-object-storage
go build
cf install-plugin -f cf-object-storage
