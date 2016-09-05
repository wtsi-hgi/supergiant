[![GoReportCard Widget]][GoReportCard] [![GoDoc Widget]][GoDoc] [![Travis Widget]][Travis]

<!-- [![Coverage Status Widget]][Coverage Status] -->

[GoReportCard Widget]: https://goreportcard.com/badge/github.com/supergiant/supergiant
[GoReportCard]: https://goreportcard.com/report/github.com/supergiant/supergiant
[GoDoc]: https://godoc.org/github.com/supergiant/supergiant
[GoDoc Widget]: https://godoc.org/github.com/supergiant/supergiant?status.svg
[Travis]: https://travis-ci.org/supergiant/supergiant
[Travis Widget]: https://travis-ci.org/supergiant/supergiant.svg?branch=master
<!-- [Coverage Status]: https://coveralls.io/github/supergiant/supergiant?branch=master
[Coverage Status Widget]: https://coveralls.io/repos/github/supergiant/supergiant/badge.svg?branch=master -->

# Supergiant

Supergiant is API-based stateful container orchestration disguised as a
developer-friendly application platform. It is based on
[Kubernetes](https://github.com/kubernetes/kubernetes).

[supergiant.io](https://supergiant.io)

### Concepts

See the [docs](docs/v0/apps.md).

### Development

Create Admin User
```shell
godep go run cmd/generate_admin_user/generate_admin_user.go --config-file config/config.json
```

Run
```shell
godep go run main.go --config-file config/config.json
```

Test
```shell
godep go test -v ./test/...
```

### License

This software is licensed under the Apache License, version 2 ("ALv2"), quoted below.

Copyright 2016 Qbox, Inc., a Delaware corporation. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License"); you may not
use this file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations under
the License.
