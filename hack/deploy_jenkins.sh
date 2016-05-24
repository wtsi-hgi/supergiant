curl -XPOST localhost:8080/v0/entrypoints -d '{
  "domain": "example.com"
}' || true

curl -XPOST localhost:8080/v0/apps -d '{
  "name": "jenkins"
}'

curl -XPOST localhost:8080/v0/apps/jenkins/components -d '{
  "name": "jenkins"
}'

curl -XPOST localhost:8080/v0/apps/jenkins/components/jenkins/releases -d '{
  "containers": [
    {
      "image": "jenkins",
      "cpu": {
        "min": 0,
        "max": 0
      },
      "ram": {
        "min": 0,
        "max": 0
      },
      "ports": [
        {
          "protocol": "HTTP",
          "number": 8080,
          "external_number": 80,
          "public": true,
          "entrypoint_domain": "example.com"
        }
      ]
    }
  ]
}'

curl -XPOST localhost:8080/v0/apps/jenkins/components/jenkins/deploy
