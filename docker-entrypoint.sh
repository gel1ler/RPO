#!/bin/sh
set -e
/usr/local/bin/server &
exec nginx -g "daemon off;"
