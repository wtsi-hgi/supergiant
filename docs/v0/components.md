# Components

A Component is a child of an App. In a sense, Components are the real building
block of Supergiant. It is synonymous with the concept of a microservice, in that
a Component should ideally have one role, or responsibility within an App.

As a basic example: within an App named "wordpress-production", there might be
Components named "mysql" and "wordpress".

#### Examples

- my-mysql-db
- wordpress
- my-redis-cache
- my-bg-service (unclustered set of workers)
- mongo-shard-1-rs (a shard of a MongoDB cluster as a replica set)
- elasticsearch (a multi-node ES cluster)

## Design

A Component does not correspond directly to a single Kubernetes resource, in the
same way as an App (which corresponds to a Namespace).

Instead a Component groups (indirectly, by way of Releases > Instances)
[Secrets](https://github.com/kubernetes/kubernetes/blob/master/docs/design/secrets.md),
[Services](https://github.com/kubernetes/kubernetes/blob/master/docs/design/services.md),
[ReplicationControllers](https://github.com/kubernetes/kubernetes/blob/master/docs/user-guide/replication-controller.md),
and [Pods](https://github.com/kubernetes/kubernetes/blob/master/docs/user-guide/pods.md).

You'll notice in the schema that a Component does not appear to hold enough
configuration to represent a deployment of containers. Other than it's own name
(which is used for external asset naming), a Component can only define a
[custom_deploy_script](custom-deploy-scripts.md), which says _how_ to deploy
whatever configuration the Component _is pointing to_.

As hinted by the phrase _"is pointing to"_, a Component *is a pointer to the
most recent configuration*, which is stored in a [Release](releases). More
specifically, a Component has 2 pointers to Releases, a current and a target
Release.

```
Component
  |
  V
  V (sorted chronologically)
  V
  |
  |-- Release
  |-- Release
  |-- [ current ] Release  (the config representing the current, or live, Instances of the Component)
  |                 |
  |                 |- Instance (spinning down)
  |                 |- Instance (active)
  |                 '- Instance (active)
  |
  |-- [ target ]  Release  (the config being applied to produce the new Instances, and decommission the current)
  |                 |
  |                 |- Instance (spinning up)
  |                 |- Instance (inactive)
  |                 '- Instance (inactive)
  |
  V
```

Before the first [deploy](releases.md#deploying) of a Component, there can only be a
_target_ Release. During subsequent deployments, there will be *both* _current_
and _target_ Releases (with target representing the newest, and soon-to-be
current). After deploying, a Component has only a _current_ Release.

#### Relations

- _belongs to_ an [App](apps.md)
- _has many_ [Releases](releases.md)

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

#### Schema

```json
{
  "name": "my-component",
  "custom_deploy_script": {
    "image": "supergiant/deploy-elasticsearch",
    "command": [
      "--app-name", "my-app", "--component-name", "my-component"
    ]
  },
  "tags": {
    "some": "tag"
  }
}
```

[Component API docs](http://supergiant-batman-364753107.us-east-1.elb.amazonaws.com:31590/docs/#/Components)
<br>
_The definition of the model and all the attributes can be found by clicking on
an operation, and then click on "Model", which is to the left of "Example Value"._
