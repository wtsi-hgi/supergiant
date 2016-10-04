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
: ${ENTRYPOINT_NAME?"Need to set ENTRYPOINT_NAME"}

# Service
cat <<EOF | supergiant kube_resources create -f -
{
  "kube_name": "$KUBE_NAME",
  "namespace": "my-couchbase",
  "kind": "Service",
  "name": "couchbase-0",
  "template": {
    "spec": {
      "type": "NodePort",
      "selector": {
        "service": "couchbase"
      },
      "ports": [
        {
          "name": "web",
          "port": 8091,
          "SUPERGIANT_ENTRYPOINT_LISTENER": {
            "entrypoint_name": "$ENTRYPOINT_NAME",
            "entrypoint_port": 8091
          }
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
  "namespace": "my-couchbase",
  "kind": "Pod",
  "name": "couchbase-0",
  "template": {
    "metadata": {
      "labels": {
        "service": "couchbase"
      }
    },
    "spec": {
      "containers": [
        {
          "name": "couchbase",
          "image": "couchbase",
          "resources": {
            "requests": {
              "cpu": 0,
              "memory": "0.5Gi"
            },
            "limits": {
              "cpu": 0.5,
              "memory": "1Gi"
            }
          },
          "volumeMounts": [
            {
              "name": "couchbase-data",
              "mountPath": "/opt/couchbase/var"
            }
          ]
        }
      ],
      "volumes": [
        {
          "name": "couchbase-data",
          "SUPERGIANT_EXTERNAL_VOLUME": {
            "type": "gp2",
            "size": 10
          }
        }
      ]
    }
  }
}
EOF

echo "Waiting for Pod to start"
while [[ $(supergiant kube_resources list --filter=kind:Pod --filter=name:couchbase-0 --format='{{ .Started }}') == 'false' ]]; do
  printf .
  sleep 1
done

# Get external address of first node
first_node_address=$(supergiant entrypoints list --filter=name:$ENTRYPOINT_NAME --format='{{ .Address }}:8091')

echo "You can reach Couchbase at $first_node_address"

kube_id=$(supergiant kubes list --filter=name:$KUBE_NAME --format='{{ .ID }}')
echo "And you can view the container log with:"
echo "supergiant kubectl -k $kube_id logs couchbase-0 --namespace=my-couchbase"
