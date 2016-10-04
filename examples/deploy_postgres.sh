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
  "namespace": "my-postgres",
  "kind": "Service",
  "name": "postgres",
  "template": {
    "spec": {
      "selector": {
        "service": "postgres"
      },
      "ports": [
        {
          "port": 5432
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
  "namespace": "my-postgres",
  "kind": "Pod",
  "name": "postgres",
  "template": {
    "metadata": {
      "labels": {
        "service": "postgres"
      }
    },
    "spec": {
      "containers": [
        {
          "name": "postgres",
          "image": "postgres",
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
          "env": [
            {
              "name": "PGDATA",
              "value": "/var/lib/postgresql/data/pgdata"
            }
          ],
          "volumeMounts": [
            {
              "name": "postgres-data",
              "mountPath": "/var/lib/postgresql/data"
            }
          ]
        }
      ],
      "volumes": [
        {
          "name": "postgres-data",
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
while [[ $(supergiant kube_resources list --filter=kind:Pod --filter=name:postgres --format='{{ .Started }}') == 'false' ]]; do
  printf .
  sleep 1
done
echo ""

echo "You can reach Postgres internally (from other containers) at postgres.my-postgres.svc.cluster.local:5432"

kube_id=$(supergiant kubes list --filter=name:$KUBE_NAME --format='{{ .ID }}')
echo "And you can view the container log with:"
echo "supergiant kubectl -k $kube_id logs postgres --namespace=my-postgres"
