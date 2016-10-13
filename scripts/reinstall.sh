#!/bin/sh
cf uninstall-plugin ObjectStorageLargeObjects
rm cf-large-objects
go build
cf install-plugin -f cf-large-objects 
