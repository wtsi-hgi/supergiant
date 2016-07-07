# Entrypoints

Entrypoints allow for Components to be reached through a public, internet-facing
address. More technically, Entrypoints are how Supergiant handles external load
balancing. While Kubernetes handles _internal_ load balancing among containers
brilliantly, we believe Entrypoints are a more efficient system for _external_
load balancing among [Nodes](nodes.md).

With vanilla Kubernetes, a Service with type set to `LoadBalancer` will create
a new AWS ELB (elastic load balancer). ELBs are a billable resource with
low per-region limits, so this type of 1-to-1 allocation of ELBs for each
Service defining external ports can be somewhat overkill.

Due to needs at qbox.io, we created Entrypoints as a way to have _shareable_
external load balancers. Each port defined still must have a unique number
(in other words you can't share port 80, but you can share the hostname), but
it can help enormously when allocating many non-standard ports on 1 ELB. We will
hopefully allow for sharing port numbers with an upcoming proxy feature.

## Design

An Entrypoint maps directly to an AWS ELB (Amazon Web Services Elastic Load
Balancer). In order to use an Entrypoint, a [Release](releases.md) must define
Ports referencing it with the field `entrypoint_domain`.

When a Release with a configured Entrypoint is deployed, the relevant port
configuration is reflected on the ELB.

Note: `domain` is simply the human-readable identifier for an Entrypoint. There
is not _yet_ DNS integration, but there most likely will be in the near future.

#### Relations

- referenced by [Releases](releases.md)

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
  "domain": "example.com"
}
```

[Entrypoint API docs](http://supergiant-batman-364753107.us-east-1.elb.amazonaws.com:31590/docs/#/Entrypoints)
<br>
_The definition of the model and all the attributes can be found by clicking on
an operation, and then click on "Model", which is to the left of "Example Value"._
