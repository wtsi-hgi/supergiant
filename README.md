# Supergiant Core

### Tests

### Documentation

```shell
godoc -http=:6060
```




- there is a Controller (Listener / Receiver) which listens for commands, and distributes to async Deployer(s)

- there is a Deployer

- there are DeployStrategies

- there is a Provisioner (which creates Kubernetes resources)

- A Service should hold reference to a Translation (A ServiceTranslation ??)



---- or is it like
```ruby
deployment = Deployment.new(service)
deployment.deployer
deployment.provisioner
```
........ but the provisioner can run with no knowledge of a deploy strategy...





A Service.............................

- belongs_to an Environment (Kube namespace)

- has_many Nodes

A Node.............................

- has_many Volumes

- has_many Containers

- shutdown_grace_period

A Container.............................

- has an Image

- has a CPU ResourceAllocation

- has a RAM ResourceAllocation

- has_many MountedVolumes (referenced from its Node)

- has_many Ports

- has a Command

- has many EnvironmentVariables

An Image.............................

A ResourceAllocation.............................

A Volume.............................

A MountedVolume.............................

A Port.............................
