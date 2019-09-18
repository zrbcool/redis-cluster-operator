### RedisCluster Status
clusterStatus  

| ClusterStatus    | Condition  | Description  |
|  ----  | ----  | ----  |
| INIT  |  | 初始状态 |
| READY  | ClusterStatus = INIT && Desired = Ready | 节点就绪未创建集群 |
| CLUSTER-OK  | Desired = Ready<br>redis-cli cluster info returns OK  |  |
| CLUSTER-CREATING  | Desired = Ready<br>redis-cli cluster info returns NOT INIT  |  |
| CLUSTER-FAIL  | Desired = Ready<br>redis-cli cluster info returns FAIL  |  |



https://github.com/antirez/redis/issues/2527  
http://download.redis.io/redis-stable/redis.conf  
```bash
########################## CLUSTER DOCKER/NAT support  ########################

# In certain deployments, Redis Cluster nodes address discovery fails, because
# addresses are NAT-ted or because ports are forwarded (the typical case is
# Docker and other containers).
#
# In order to make Redis Cluster working in such environments, a static
# configuration where each node knows its public address is needed. The
# following two options are used for this scope, and are:
#
# * cluster-announce-ip
# * cluster-announce-port
# * cluster-announce-bus-port
#
# Each instruct the node about its address, client port, and cluster message
# bus port. The information is then published in the header of the bus packets
# so that other nodes will be able to correctly map the address of the node
# publishing the information.
#
# If the above options are not used, the normal Redis Cluster auto-detection
# will be used instead.
#
# Note that when remapped, the bus port may not be at the fixed offset of
# clients port + 10000, so you can specify any port and bus-port depending
# on how they get remapped. If the bus-port is not set, a fixed offset of
# 10000 will be used as usually.
#
# Example:
#
# cluster-announce-ip 10.1.1.5
# cluster-announce-port 6379
# cluster-announce-bus-port 6380
```

sed -i 's|${cluster-announce-ip}|127.0.0.1|g' redis.conf.test  
sed -i 's|${cluster-announce-port}|7000|g' redis.conf.test  
sed -i 's|${cluster-announce-bus-port}|17000|g' redis.conf.test  

/usr/local/bin/docker-entrypoint.sh
https://github.com/docker-library/redis/blob/master/5.0/Dockerfile

docker run --rm --name redis -e CLUSTER_ANNOUNCE_IP=192.168.100.202 zrbcool/redis:5.0.5 redis-server /data/redis.conf  
