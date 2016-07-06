# Apps

An App is what groups Components. You could also say an App is how to organize
some collection of (micro)services. In practical terms, it often serves as an
environment, such as "my-app-production", but organization is flexible, and up
to the user.

There may be situations where one App stores a database Component, and a
Component running application code. However, there may also be situations where
using an App to represent one large system component (or "multi-component")
makes sense (such as having a Component for each shard of a large Mongo cluster).

#### Examples

- my-website
- my-website-production
- my-website-staging
- my-huge-mongo-cluster
- my-saas-application

## Design

In technical terms, an App is nothing more than a Kubernetes
[Namespace](https://github.com/kubernetes/kubernetes/blob/master/docs/admin/namespaces.md)
with a different name.

In the context of Supergiant, Apps are the top-level resource for all deployments,
that is to say it is the parent of the relational chain
[Components](components.md) > [Releases](releases.md) > [Instances](instances.md).

#### Relations

An App _has many_ [Components](components.md).

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
  "name": "my-app",
  "tags": {
    "some": "tag"
  }
}
```

[App API docs](http://supergiant-batman-364753107.us-east-1.elb.amazonaws.com:31590/docs/#/Apps)
<br>
_The definition of the model and all the attributes can be found by clicking on
an operation, and then click on "Model", which is to the left of "Example Value"._
