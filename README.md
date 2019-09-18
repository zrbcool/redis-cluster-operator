### RedisCluster Status
clusterStatus  

| ClusterStatus    | Condition  | Description  |
|  ----  | ----  | ----  |
| INIT  |  | 初始状态 |
| READY  | ClusterStatus = INIT && Desired = Ready | 节点就绪未创建集群 |
| CLUSTER-OK  | Desired = Ready<br>redis-cli cluster info returns OK  |  |
| CLUSTER-CREATING  | Desired = Ready<br>redis-cli cluster info returns NOT INIT  |  |
| CLUSTER-FAIL  | Desired = Ready<br>redis-cli cluster info returns FAIL  |  |

