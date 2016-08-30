set -e

echo "creating entrypoint"
entrypoint_id=$(curl -s -XPOST http://admin:password@localhost:8080/api/v0/entrypoints -d "{
  \"kube_id\": $KUBE_ID,
  \"name\": \"elasticsearch\"
}" | grep -Eo '"id": \d+' | head -1 | awk '{ print $2 }')
sleep 5 # wait for entrypoint address

echo "creating app"
app_id=$(curl -s -XPOST http://admin:password@localhost:8080/api/v0/apps -d "{
  \"kube_id\": $KUBE_ID,
  \"name\": \"elasticsearch\"
}" | grep -Eo '"id": \d+' | head -1 | awk '{ print $2 }')

echo "creating component"
component_id=$(curl -s -XPOST http://admin:password@localhost:8080/api/v0/components -d "{
  \"app_id\": $app_id,
  \"name\": \"elasticsearch\"
}" | grep -Eo '"id": \d+' | head -1 | awk '{ print $2 }')

echo "creating release"
curl -s -XPOST http://admin:password@localhost:8080/api/v0/releases -d "{
  \"component_id\": $component_id,

  \"instance_count\": 3,

  \"config\": {
    \"termination_grace_period\": 10,
    \"volumes\": [
      {
        \"name\": \"elasticsearch-data-1\",
        \"type\": \"gp2\",
        \"size\": 10
      },
      {
        \"name\": \"elasticsearch-data-2\",
        \"type\": \"gp2\",
        \"size\": 10
      }
    ],
    \"containers\": [
      {
        \"image\": \"elasticsearch\",

        \"command\": [
          \"elasticsearch\",
          \"-Des.insecure.allow.root=true\",
          \"-Des.discovery.zen.ping.multicast.enabled=false\",
          \"-Des.bootstrap.mlockall=true\",
          \"-Des.index.number_of_shards=5\",
          \"-Des.index.number_of_replicas=1\",
          \"-Des.discovery.zen.ping.unicast.hosts=elasticsearch.elasticsearch.svc.cluster.local:9300\",
          \"-Des.path.data=/data-1\",
          \"-Des.path.logs=/data-1\",
          \"-Des.processors=1\",
          \"-Des.discovery.zen.minimum_master_nodes=2\"
        ],

        \"cpu_request\": 0,
        \"cpu_limit\": 0.5,

        \"ram_request\": \"1.5Gi\",
        \"ram_limit\": \"2Gi\",

        \"mounts\": [
          {
            \"volume\": \"elasticsearch-data-1\",
            \"path\": \"/data-1\"
          },
          {
            \"volume\": \"elasticsearch-data-2\",
            \"path\": \"/data-2\"
          }
        ],
        \"ports\": [
          {
            \"number\": 9200,
            \"external_number\": 9200,
            \"public\": true,
            \"entrypoint_id\": $entrypoint_id
          },
          {
            \"number\": 9300
          }
        ]
      }
    ]
  }
}"

echo "deploying"
curl -XPOST http://admin:password@localhost:8080/api/v0/components/$component_id/deploy

# curl -s -XPOST http://admin:password@localhost:8080/api/v0/releases -d "{
#   \"component_id\": $component_id,
#   \"instance_count\": 2
# }"
#
# curl -XPOST http://admin:password@localhost:8080/api/v0/components/$component_id/deploy
