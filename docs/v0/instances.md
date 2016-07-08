# Instances

An Instance is an abstraction around a running (or soon-to-be-running)
_container_. It conveniently bundles logic for dealing with the container,
its parent K8S RC+Pod, K8S Services (if there are per_instance ports), and
mountable cloud volumes.

When you define the configuration of a Component in a [Release](releases.md),
portions such as volumes, CPU/RAM allocation, and per_instance ports affect
resources that are provisioned for each instance. For example, a Release with 1
volume defined, 1 container defined, and ram.min set to 1Gi will result in
allocating 3 containers, each with its own volume attached (3 total), and each
with 1Gi RAM reserved (3Gi total).

## Design

Each Instance of a Release is always responsible for a Kubernetes
ReplicationController with an attached Pod, which houses running container(s).
If volume(s) are defined, each Instance will either create (if first deploy) or
detach/reattach (if subsequent deploy) a volume reflecting the Release
definition.

If port(s) are defined with `per_instance` set to true, then each Instance also
has a respective Kubernetes Service provisioned.

During a deployment (following the first), Instances belonging to a Component's
_current_ Release are *stopped*, which deletes the RC+Pod and detaches any
volumes. Instances of the _target_ Release are then *started*, which will create
the RC+Pod with the new configuration and reattach the volume.

If volumes are being resized (as specified by the user), then they will be
resized before reattaching.

See [deploy/deploy.go](https://github.com/supergiant/supergiant/tree/master/deploy/deploy.go) to see the full internal deploy
process. It utilizes the API client, meaning this logic can be used in
user-built [custom_deploy_scripts](custom-deploy-scripts.md).

#### Relations

- _belongs to_ a [Release](releases.md)

```
Registries
 |
 '- Repos <---------.
                    |
Entrypoints <-------|
                    \
Apps                 \
 |                    \
 '- Components         |
      |                |
      '- Releases -->--'
           |
         (current or target)
           |
           '- Instances
```

[Instance API docs](http://swagger.supergiant.io/docs/#/Instances)
<br>
_The definition of the model and all the attributes can be found by clicking on
an operation, and then click on "Model", which is to the left of "Example Value"._
