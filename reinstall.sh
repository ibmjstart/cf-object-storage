#!/bin/sh
cf uninstall-plugin LargeObjectsPlugin
rm cf-large-objects
go build
cf install-plugin -f cf-large-objects 
