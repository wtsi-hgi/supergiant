################################################################################
# NOTE This is using cmd/cli/cli.go, built as a `supergiant` binary, which can #
#      be done using the following command:                                    #
#                                                                              #
#      go build -o $GOPATH/bin/supergiant cmd/cli/cli.go                       #
#                                                                              #
# You'll then need to globally configure server address and API token:         #
#                                                                              #
#      supergiant configure -s <server_address> -t <api_token>                 #
#                                                                              #
################################################################################

set -e

: ${KUBE_NAME?"Need to set KUBE_NAME"}

# Service
cat <<EOF | supergiant kube_resources create -f -
{
  "kube_name": "$KUBE_NAME",
  "namespace": "my-redis",
  "kind": "Service",
  "name": "redis",
  "template": {
    "spec": {
      "selector": {
        "service": "redis"
      },
      "ports": [
        {
          "port": 6379
        }
      ]
    }
  }
}
EOF

# Pod
cat <<EOF | supergiant kube_resources create -f -
{
  "kube_name": "$KUBE_NAME",
  "namespace": "my-redis",
  "kind": "Pod",
  "name": "redis",
  "template": {
    "metadata": {
      "labels": {
        "service": "redis"
      }
    },
    "spec": {
      "containers": [
        {
          "name": "redis",
          "image": "redis:alpine",
          "command": [
            "redis-server", "--appendonly", "yes"
          ],
          "resources": {
            "requests": {
              "cpu": 0,
              "memory": 0
            },
            "limits": {
              "cpu": 0.5,
              "memory": "2Gi"
            }
          },
          "volumeMounts": [
            {
              "name": "redis-data",
              "mountPath": "/data"
            }
          ]
        }
      ],
      "volumes": [
        {
          "name": "redis-data",
          "SUPERGIANT_EXTERNAL_VOLUME": {
            "size": 20
          }
        }
      ]
    }
  }
}
EOF

echo "Waiting for Pod to start"
while [[ $(supergiant kube_resources list --filter=kind:Pod --filter=name:redis --format='{{ .Started }}') == 'false' ]]; do
  printf .
  sleep 1
done

echo "You can reach Redis internally (from other containers) at redis.my-redis.svc.cluster.local:6379"

kube_id=$(supergiant kubes list --filter=name:$KUBE_NAME --format='{{ .ID }}')
echo "And you can view the container log with:"
echo "supergiant kubectl -k $kube_id logs redis --namespace=my-redis"
