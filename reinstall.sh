#!/bin/sh
cf uninstall-plugin LargeObjectsPlugin
go build
cf install-plugin -f cf-large-objects 
