#!/bin/sh
cf uninstall-plugin cf-object-storage
rm cf-object-storage

if [[ "$1" == "-a" ]]; then
	go build -a
else
	go build
fi

cf install-plugin -f cf-object-storage
