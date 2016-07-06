# Registries

A Registry is the parent of [Repos](repos.md) where private container images are
stored. Supergiant currently has support for Dockerhub (as far as it has been
tested). Repos must be created on the Dockerhub Registry in order to provision
containers from private images.

## Design

#### Relations

- _has many_ [Repos](repos.md)

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
  "name": "dockerhub"
}
```
