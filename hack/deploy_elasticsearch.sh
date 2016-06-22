curl -XPOST $HOST/v0/registries/dockerhub/repos -d '{
  "name": "qbox",
  "key": "'$QBOX_DOCKERHUB_KEY'"
}' || true

curl -XPOST $HOST/v0/entrypoints -d '{
  "domain": "example.com"
}' || true

curl -XPOST $HOST/v0/apps -d '{
  "name": "test"
}'

curl -XPOST $HOST/v0/apps/test/components -d '{
  "name": "elasticsearch",
  "custom_deploy_script": {
    "image": "supergiant/deploy-elasticsearch:latest",
    "command": [
      "/deploy-elasticsearch",
      "--app-name",
      "test",
      "--component-name",
      "elasticsearch"
    ]
  }
}'

curl -XPOST $HOST/v0/apps/test/components/elasticsearch/releases -d '{
  "instance_count": 1,
  "termination_grace_period": 10,
  "volumes": [
    {
      "name": "elasticsearch-data",
      "type": "gp2",
      "size": 10
    }
  ],
  "containers": [
    {
      "image": "qbox/qbox-docker:2.1.1",
      "cpu": {
        "min": 0,
        "max": 0.25
      },
      "ram": {
        "min": "1Gi",
        "max": "1Gi"
      },
      "mounts": [
        {
          "volume": "elasticsearch-data",
          "path": "/data-1"
        }
      ],
      "ports": [
        {
          "protocol": "HTTP",
          "number": 9200,
          "external_number": 9200,
          "public": true,
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
          "value": "/data-1"
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
          "value": "512m"
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

curl -XPOST $HOST/v0/apps/test/components/elasticsearch/deploy
