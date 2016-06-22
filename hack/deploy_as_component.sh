curl -XPOST localhost:8080/v0/entrypoints -d '{
  "domain": "example.com"
}' || true

curl -XPOST localhost:8080/v0/apps -d '{
  "name": "sg-test"
}'

curl -XPOST localhost:8080/v0/apps/sg-test/components -d '{
  "name": "api"
}'

curl -XPOST localhost:8080/v0/apps/sg-test/components/api/releases -d '{
  "instance_count": 1,
  "containers": [
    {
      "name": "api",
      "image": "supergiant/supergiant-api:unstable-internal_addrs_for_external_ports",
      "command": [
        "/supergiant-api",
        "--etcd-hosts",
        "http://localhost:2379",
        "--ar",
        "'$AR'",
        "--az",
        "'$AZ'",
        "--sg",
        "'$SG'",
        "--sid",
        "'$SID'",
        "--kh",
        "'$KH'",
        "--k8s-insecure-https",
        "--ku",
        "'$KU'",
        "--kp",
        "'$KP'",
        "--aws-access-key",
        "'$AWS_ACCESS_KEY'",
        "--aws-secret-key",
        "'$AWS_SECRET_KEY'"
      ],
      "ports": [
        {
          "protocol": "HTTP",
          "number": 8080,
          "external_number": 40000,
          "public": true,
          "entrypoint_domain": "example.com"
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

curl -XPOST localhost:8080/v0/apps/sg-test/components/api/deploy
