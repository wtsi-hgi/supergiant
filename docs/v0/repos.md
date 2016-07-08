# Repos

_v1.0 is likely to change the name of Repos to Organizations, since that is
both technically and more intuitively correct._

A Repo is what stores secret authentication data for private organizations on
a hosted docker-registry. Currently, Supergiant only supports Dockerhub as a
container image [Registry](registries.md) (we would love PRs to support more).

## Design

If you have a private container image hosted on Dockerhub, referenced by
"my_company/my_app:1.0.0", then the `name` of the Repo would be *my_company*.
The `key` would be the base64-encoded combination of Dockerhub credentials.

To be used, a Repo must be referenced in the `image` field of a container in a
[Release](releases.md), the value being something like "my_company/my_app:1.0.0".
Once a Release is deployed, a Kubernetes [Secret](https://github.com/kubernetes/kubernetes/blob/master/docs/design/secrets.md)
is created in the relevant App's Namespace to allow a container to be provisioned
from the private image.

#### Relations

- _belongs to_ a [Registry](registries.md)
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
  "name": "my_company",
  "key": "base64-credential-data"
}
```

[Repo API docs](http://swagger.supergiant.io/docs/#/Repos_(Dockerhub_orgs))
<br>
_The definition of the model and all the attributes can be found by clicking on
an operation, and then click on "Model", which is to the left of "Example Value"._
