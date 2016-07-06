# Custom Deploy Scripts

A `custom_deploy_script` can be defined on a [Component](components.md) to
extend or modify the standard Supergiant deployment logic.

The standard deployment process that Supergiant utilizes is API-based, and can
be seen at [deploy/deploy.go](https://github.com/supergiant/supergiant/tree/master/deploy/deploy.go).
Notice how our Elasticsearch custom_deploy_script,
[supergiant/deploy-elasticsearch](https://github.com/supergiant/deploy-elasticsearch)
_(see pkg/deploy.go)_, simply copies the the standard Supergiant deploy.go code
and adds Elasticsearch deployment logic in particular spots.

Custom Deploy Scripts _must_ be defined as a runnable container. You can see how
we reference our supergiant/deploy-elasticsearch in
[hack/deploy_elasticsearch.sh](https://github.com/supergiant/supergiant/tree/master/hack/deploy_elasticsearch.sh).
