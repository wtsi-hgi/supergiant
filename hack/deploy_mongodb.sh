curl -XPOST localhost:8080/v0/entrypoints -d '{
  "domain": "example.com"
}' || true

curl -XPOST localhost:8080/v0/apps -d '{
  "name": "test"
}'

curl -XPOST localhost:8080/v0/apps/test/components -d '{
  "name": "mongo",
  "custom_deploy_script": {
    "image": "supergiant/deploy-mongodb:latest",
    "command": [
      "/deploy-mongodb",
      "--app-name",
      "test",
      "--component-name",
      "mongo"
    ]
  }
}'

curl -XPOST localhost:8080/v0/apps/test/components/mongo/releases -d '{
  "instance_count": 3,
  "volumes": [
    {
      "name": "mongo-data",
      "type": "gp2",
      "size": 10
    }
  ],
  "containers": [
    {
      "image": "mongo",
      "command": [
        "mongod",
        "--replSet",
        "rs0"
      ],
      "cpu": {
        "max": 0.25
      },
      "ram": {
        "max": "1Gi"
      },
      "mounts": [
        {
          "volume": "mongo-data",
          "path": "/data/db"
        }
      ],
      "ports": [
        {
          "protocol": "TCP",
          "number": 27017,
          "public": true,
          "per_instance": true,
          "entrypoint_domain": "example.com"
        }
      ]
    }
  ]
}'

curl -XPOST localhost:8080/v0/apps/test/components/mongo/deploy
