set -e

echo "creating entrypoint"
entrypoint_id=$(curl -s -XPOST http://admin:password@localhost:8080/api/v0/entrypoints -d '{
  "kube_id": 8,
  "name": "jenkins"
}' | grep -Eo '"id": \d+' | head -1 | awk '{ print $2 }')

echo "creating app"
app_id=$(curl -s -XPOST http://admin:password@localhost:8080/api/v0/apps -d '{
  "kube_id": 8,
  "name": "jenkins"
}' | grep -Eo '"id": \d+' | head -1 | awk '{ print $2 }')

echo "creating component"
component_id=$(curl -s -XPOST http://admin:password@localhost:8080/api/v0/components -d "{
  \"app_id\": $app_id,
  \"name\": \"jenkins\"
}" | grep -Eo '"id": \d+' | head -1 | awk '{ print $2 }')

echo "creating release"
curl -s -XPOST http://admin:password@localhost:8080/api/v0/releases -d "{
  \"component_id\": $component_id,

  \"config\": {

    \"volumes\": [
      {
        \"name\": \"test\",
        \"type\": \"gp2\",
        \"size\": 10
      }
    ],

    \"containers\": [
      {
        \"image\": \"jenkins\",

        \"mounts\": [
          {
            \"volume\": \"test\",
            \"path\": \"/mnt\"
          }
        ],

        \"ports\": [
          {
            \"protocol\": \"HTTP\",
            \"number\": 8080,
            \"external_number\": 80,
            \"public\": true,
            \"entrypoint_id\": $entrypoint_id
          }
        ]
      }
    ]
  }
}"

echo "deploying"
curl -XPOST http://admin:password@localhost:8080/api/v0/components/$component_id/deploy
