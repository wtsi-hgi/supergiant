set -e

curl -XPOST localhost:8080/v0/registries/dockerhub/repos -d '{
  "name": "supergiant",
  "key": "'$SUPERGIANT_DOCKERHUB_KEY'"
}' || true

curl -XPOST localhost:8080/v0/entrypoints -d '{
  "domain": "example.com"
}' || true

curl -XPOST localhost:8080/v0/apps -d '{
  "name": "supergiant-io"
}'

curl -XPOST localhost:8080/v0/apps/supergiant-io/components -d '{
  "name": "mysql"
}'

curl -XPOST localhost:8080/v0/apps/supergiant-io/components/mysql/releases -d '{
  "volumes": [
    {
      "name": "data",
      "type": "gp2",
      "size": 20
    }
  ],
  "containers": [
    {
      "image": "mysql:5.6.30",
      "ports": [
        {
          "protocol": "TCP",
          "number": 3306
        }
      ],
      "mounts": [
        {
          "volume": "data",
          "path": "/var/lib/mysql"
        }
      ],
      "env": [
        {
          "name": "MYSQL_DATABASE",
          "value": "'$MYSQL_DATABASE'"
        },
        {
          "name": "MYSQL_ROOT_PASSWORD",
          "value": "'$MYSQL_ROOT_PASSWORD'"
        },
        {
          "name": "MYSQL_USER",
          "value": "'$MYSQL_USER'"
        },
        {
          "name": "MYSQL_PASSWORD",
          "value": "'$MYSQL_PASSWORD'"
        }
      ]
    }
  ]
}'

curl -XPOST localhost:8080/v0/apps/supergiant-io/components/mysql/deploy

curl -XPOST localhost:8080/v0/apps/supergiant-io/components -d '{
  "name": "craft"
}'

curl -XPOST localhost:8080/v0/apps/supergiant-io/components/craft/releases -d '{
  "containers": [
    {
      "image": "supergiant/supergiant-cms:docker",
      "ports": [
        {
          "protocol": "HTTP",
          "number": 80,
          "external_number": 80,
          "public": true,
          "entrypoint_domain": "example.com"
        }
      ],
      "env": [
        {
          "name": "CRAFT_DB_SERVER",
          "value": "mysql.supergiant-io.svc.cluster.local"
        },
        {
          "name": "CRAFT_DB_USER",
          "value": "'$MYSQL_USER'"
        },
        {
          "name": "CRAFT_DB_PASSWORD",
          "value": "'$MYSQL_PASSWORD'"
        },
        {
          "name": "CRAFT_DB_NAME",
          "value": "'$MYSQL_DATABASE'"
        }
      ]
    }
  ]
}'

curl -XPOST localhost:8080/v0/apps/supergiant-io/components/craft/deploy
