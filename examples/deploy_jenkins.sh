set -e

echo "creating app"
app_id=$(curl -H "Authorization: SGAPI token=\"$API_TOKEN\"" -s -XPOST https://localhost:8081/api/v0/apps -d "{
  \"kube_id\": $KUBE_ID,
  \"name\": \"jenkins\"
}" | grep -Eo '"id": \d+' | head -1 | awk '{ print $2 }')

echo "creating component"
component_id=$(curl -H "Authorization: SGAPI token=\"$API_TOKEN\"" -s -XPOST https://localhost:8081/api/v0/components -d "{
  \"app_id\": $app_id,
  \"name\": \"jenkins\"
}" | grep -Eo '"id": \d+' | head -1 | awk '{ print $2 }')

echo "creating release"
curl -H "Authorization: SGAPI token=\"$API_TOKEN\"" -s -XPOST https://localhost:8081/api/v0/releases -d "{
  \"component_id\": $component_id,
  \"config\": {
    \"containers\": [
      {
        \"image\": \"jenkins\",
        \"ports\": [
          {
            \"protocol\": \"HTTP\",
            \"number\": 8080,
            \"external_number\": 80,
            \"public\": true
          }
        ]
      }
    ]
  }
}"

echo "deploying"
curl -H "Authorization: SGAPI token=\"$API_TOKEN\"" -s -XPOST https://localhost:8081/api/v0/components/$component_id/deploy
