curl -XPOST localhost:8080/v0/entrypoints -d '{
  "domain": "test.supergiant.io"
}' || true

curl -XPOST localhost:8080/v0/apps -d '{
  "name": "supergiant"
}'

curl -XPOST localhost:8080/v0/apps/supergiant/components -d '{
  "name": "api"
}'

curl -XPOST localhost:8080/v0/apps/supergiant/components/api/releases -d '{
  "instance_count": 1,
  "containers": [
    {
      "name": "api",
      "image": "supergiant/supergiant-api:latest",
      "command": [
        "/supergiant-api",
        "--etcd-host",
        "http://localhost:2379",
        "--ar",
        "us-east-1",
        "--az",
        "us-east-1c",
        "--sg",
        "sg-8923f3f1",
        "--sid",
        "subnet-21e06f57",
        "--kh",
        "52.90.24.78",
        "--https-mode",
        "--ku",
        "admin",
        "--kp",
        "86J8mb3b4bDU32cX",
        "--access-key",
        "'$AWS_ACCESS_KEY'",
        "--secret-key",
        "'$AWS_SECRET_KEY'"
      ],
      "ports": [
        {
          "protocol": "HTTP",
          "number": 8080,
          "external_number": 40000,
          "public": true,
          "entrypoint_domain": "test.supergiant.io"
        }
      ]
    },
    {
      "image": "quay.io/coreos/etcd:latest",
      "command": [
        "/etcd",
        "--name",
        "etcd",
        "--initial-advertise-peer-urls",
        "http://localhost:2380",
        "--listen-peer-urls",
        "http://0.0.0.0:2380",
        "--listen-client-urls",
        "http://0.0.0.0:2379",
        "--advertise-client-urls",
        "http://localhost:2379",
        "--initial-cluster",
        "etcd=http://localhost:2380",
        "--initial-cluster-state",
        "new"
      ],
      "ports": [
        {
          "number": 2379
        },
        {
          "number": 2380
        }
      ]
    }
  ]
}'

curl -XPOST localhost:8080/v0/apps/supergiant/components/api/deploy
