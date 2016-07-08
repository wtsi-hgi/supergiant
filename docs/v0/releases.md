# Releases

A Release is _a_ full configuration of a Component, _released_ (i.e. committed,
deployed, published, etc.) at a certain time. Releases can be verbose, as they
represent an entire topology of a Component.

## Design

The physical manifestation of a Component comes in the form of its _current_
Release. The live assets (cloud volumes, containers in Kubernetes Pods, etc.) of
a Component reflect the most recently _released configuration_ (i.e. Release).

#### Deploying

When deploying, a Component is the [union](https://en.wikipedia.org/wiki/Union_(set_theory))
of (active) Instances between its _current_ and _target_ Releases. More
specifically, [Instances](instances.md), which belong to Releases, are swapped
during the deploy (e.g. current-release-instance-0 is replaced by
target-release-instance-0). This is how configuration is _released_ across the
Instances of a Component.

_The following illustrates a Component mid-deploy, specifically focusing on the
Instance "hand-off" procedure. Kubernetes Services are updated before and after
this process (if there are port changes between Releases)._

```
Component
  |
  V
  V (sorted chronologically)
  V
  |
  |-- Release
  |-- Release
  |-- Release [ CURRENT ]
  |     |
  |     |- Instance
  |     |    |
  |     |    |- K8S RC+Pod, i.e. container(s) (spinning DOWN)
  |     |                          |
  |     |                       (mount)
  |     |                          |
  |     |                          '- cloud volume(s) --(detaching)---.
  |     |                                                             |
  |     |- Instance (active)                                          |
  |     '- Instance (active)                                          |
  |                                                                   |
  |-- Release [ TARGET ]                                              |
  |     |                                                             |
  |     |- Instance                                                   |
  |     |    |                                                        |
  |     |    |- K8S RC+Pod, i.e. container(s) (spinning UP)           |
  |     |                          |                                  |
  |     |                       (mount)                               |
  |     |                          |                                  |
  |     |                          '- cloud volume(s) <--(attaching)--'
  |     |
  |     |- Instance (inactive)
  |     '- Instance (inactive)
  |
  V
```

It is intuitive to think of a Release like a version-controlled code repository
(like a Git repo), where the "active" code must be considered at _a point in
time_. However, to note some low-level nuance, a Release is not merely a diff,
but in reality a full configuration at each step.

The implication is that, under the hood, Supergiant is indifferent to the
historical changes of each Release relative to the last. It only cares that its
_de-provisioning the current Release_ and _provisioning the target Release_, in
a predictably sequential way (instance-by-instance).

#### Relations

- _belongs to_ a [Component](components.md)
- _has many_ [Instances](instances.md)

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
  "instance_count": 3,
  "volumes": [
    {
      "name": "data",
      "type": "gp2",
      "size": 10
    }
  ],
  "containers": [
    {
      "image": "mongo",
      "command": [
        "mongod",
        "--replSet",
        "rs0"
      ],
      "cpu": {
        "max": 0.25
      },
      "ram": {
        "min": "256Mi",
        "max": "1Gi"
      },
      "mounts": [
        {
          "volume": "data",
          "path": "/data/db"
        }
      ],
      "ports": [
        {
          "protocol": "TCP",
          "number": 27017,
          "public": true,
          "per_instance": true,
          "entrypoint_domain": "example.com"
        }
      ]
    }
  ]
}
```

[Release API docs](http://swagger.supergiant.io/docs/#/Releases)
<br>
_The definition of the model and all the attributes can be found by clicking on
an operation, and then click on "Model", which is to the left of "Example Value"._
