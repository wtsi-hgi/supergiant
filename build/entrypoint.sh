#!/bin/ash
set -e
[ $DEBUG ] && set -x

cp -f /usr/share/zoneinfo/$TIMEZONE /etc/localtime

$@
