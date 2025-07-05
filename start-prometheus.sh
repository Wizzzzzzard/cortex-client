#!/usr/bin/env bash

set -e

PROM1_PORT=9090
PROM2_PORT=9091
PROM_IMAGE="prom/prometheus:latest"

start() {
	echo "Starting Prometheus instance 1 in Docker on port $PROM1_PORT..."
	CID1=$(docker run -d --rm -p $PROM1_PORT:9090 $PROM_IMAGE)
	echo $CID1 > prometheus1.cid

	echo "Starting Prometheus instance 2 in Docker on port $PROM2_PORT..."
	CID2=$(docker run -d --rm -p $PROM2_PORT:9090 $PROM_IMAGE)
	echo $CID2 > prometheus2.cid

	sleep 2
	echo "Both Prometheus containers started."
	echo "Access them at:"
	echo "  http://localhost:$PROM1_PORT"
	echo "  http://localhost:$PROM2_PORT"
}

stop() {
	for cidfile in prometheus1.cid prometheus2.cid; do
		if [ -f "$cidfile" ]; then
			cid=$(cat "$cidfile")
			if [ -n "$cid" ]; then
				echo "Stopping Prometheus container $cid..."
				docker stop "$cid" || true
				echo "Stopped."
			else
				echo "No container ID in $cidfile."
			fi
			rm -f "$cidfile"
		else
			echo "No container ID file $cidfile found."
		fi
	done
}

case "$1" in
	stop)
		stop
		;;
	start|"")
		start
		;;
	*)
		echo "Usage: $0 [start|stop]"
		exit 1
		;;
esac
