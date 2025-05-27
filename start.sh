#!/bin/sh
redis-server /etc/redis/redis.conf --daemonize yes
sleep 2
./main