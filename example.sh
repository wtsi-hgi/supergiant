curl -XPOST localhost:8080/v0/registries/dockerhub/repos -d '{
  "name": "qbox",
  "key": "'$QBOX_DOCKERHUB_KEY'"
}' || true

curl -XPOST localhost:8080/v0/entrypoints -d '{
  "domain": "example.com"
}' || true

curl -XPOST localhost:8080/v0/apps -d '{
  "name": "test"
}'

curl -XPOST localhost:8080/v0/apps/test/components -d '{
  "name": "elasticsearch"
}'

curl -XPOST localhost:8080/v0/apps/test/components/elasticsearch/releases -d '{
  "instance_count": 1,
  "termination_grace_period": 10,
  "volumes": [
    {
      "name": "elasticsearch-data-1",
      "type": "gp2",
      "size": 10
    },
    {
      "name": "elasticsearch-data-2",
      "type": "gp2",
      "size": 10
    }
  ],
  "containers": [
    {
      "image": "qbox/qbox-docker:2.1.1",
      "cpu": {
        "min": 0,
        "max": 0.5
      },
      "ram": {
        "min": "1.5Gi",
        "max": "2Gi"
      },
      "mounts": [
        {
          "volume": "elasticsearch-data-1",
          "path": "/data-1"
        },
        {
          "volume": "elasticsearch-data-2",
          "path": "/data-2"
        }
      ],
      "ports": [
        {
          "protocol": "HTTP",
          "number": 9200,
          "external_number": 33666,
          "public": true,
          "per_instance": true,
          "entrypoint_domain": "example.com"
        },
        {
          "protocol": "TCP",
          "number": 9300
        }
      ],
      "env": [
        {
          "name": "CLUSTER_ID",
          "value": "SG_TEST"
        },
        {
          "name": "NODE_NAME",
          "value": "SG_TEST_{{ instance_id }}"
        },
        {
          "name": "MASTER_ELIGIBLE",
          "value": "true"
        },
        {
          "name": "DATA_PATHS",
          "value": "/data-1,/data-2"
        },
        {
          "name": "UNICAST_HOSTS",
          "value": "elasticsearch.test.svc.cluster.local:9300"
        },
        {
          "name": "MIN_MASTER_NODES",
          "value": "1"
        },
        {
          "name": "CORES",
          "value": "1"
        },
        {
          "name": "ES_HEAP_SIZE",
          "value": "1024m"
        },
        {
          "name": "INDEX_NUMBER_OF_SHARDS",
          "value": "4"
        },
        {
          "name": "INDEX_NUMBER_OF_REPLICAS",
          "value": "0"
        }
      ]
    }
  ]
}'

curl -XPOST localhost:8080/v0/apps/test/components/elasticsearch/deploy
