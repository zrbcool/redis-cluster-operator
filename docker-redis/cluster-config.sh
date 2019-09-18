#!/bin/sh
if test -z "$CLUSTER_ANNOUNCE_IP"
then
	VAR_IP="127.0.0.1"
else
	VAR_IP=${CLUSTER_ANNOUNCE_IP}
fi

if test -z "$CLUSTER_ANNOUNCE_PORT"
then
	VAR_PORT="7000"
else
	VAR_PORT=${CLUSTER_ANNOUNCE_PORT}
fi

if test -z "$CLUSTER_ANNOUNCE_PORT"
then
        VAR_BUS_PORT="17000"
else
        VAR_BUS_PORT=${CLUSTER_ANNOUNCE_BUS_PORT}
fi

echo cluster-announce-ip=${VAR_IP}
echo cluster-announce-port=${VAR_PORT}
sed -i 's|${cluster-announce-ip}|'`echo $VAR_IP`'|g' /data/redis.conf
sed -i 's|${cluster-announce-port}|'`echo $VAR_PORT`'|g' /data/redis.conf
sed -i 's|${cluster-announce-bus-port}|'`echo $VAR_BUS_PORT`'|g' /data/redis.conf
