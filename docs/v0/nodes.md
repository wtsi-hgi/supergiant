# Nodes

A Node is a server (i.e. VM, EC2 instance). One of the main goals of Supergiant
is to abstract away server management entirely -- while there is a full CRUD API
for Nodes, the [Capacity Service](capacity-service.md) is capable of managing
servers autonomously, so a user can focus on allocating containers.

## Design

In technical terms, a Node is controllable abstraction around a cloud server and
the respective Kubernetes [Node](https://github.com/kubernetes/kubernetes/blob/master/docs/admin/node.md)
which it hosts. This allows for direct creation and deletion of Nodes, which
cannot be handled directly from the Kubernetes API.

#### Schema

```json
{
  "class": "m4.large"
}
```

[Node API docs](http://supergiant-batman-364753107.us-east-1.elb.amazonaws.com:31590/docs/#/Nodes)
<br>
_The definition of the model and all the attributes can be found by clicking on
an operation, and then click on "Model", which is to the left of "Example Value"._
