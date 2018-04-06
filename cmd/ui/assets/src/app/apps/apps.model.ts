export class AppsModel {
  app = {
    'model': {
      'chart_name': '',
      'chart_version': '',
      'config': null,
      // this needs to be called cluster in display
      'kube_name': '',
      'name': '',
      'repo_name': '',
      'namespace': ''
    },
    'schema': {
      'properties': {
        'chart_name': {
          'readonly': true,
          '$id': '/properties/chart_name',
          'type': 'string',
          'title': 'Chart Name',
          'default': '',
          'examples': [
            ''
          ]

        },
        'chart_version': {
          '$id': '/properties/chart_version',
          'type': 'string',
          'title': 'Chart Version',
          'default': '',
          'examples': [
            ''
          ]

        },
        'kube_name': {
          '$id': '/properties/kube_name',
          'type': 'string',
          'title': 'Cluster Name',
          'default': '',
          'enum': [],
          'examples': [
            ''
          ]

        },
        'name': {
          '$id': '/properties/name',
          'type': 'string',
          'title': 'Deployment Name (Optional: Randomly Generated If Not Specified.)',
          'default': '',
          'examples': [
            ''
          ]

        },
        'repo_name': {
          'readonly': true,
          '$id': '/properties/repo_name',
          'type': 'string',
          'title': 'Helm Repo Name',
          'default': '',
          'examples': [
            ''
          ]

        },
        'namespace': {
          '$id': '/properties/namespace',
          'type': 'string',
          'title': 'Namespace (Optional: The Kubernetes Namespace You Would Like To Deploy To.)',
          'default': '',
          'examples': [
            ''
          ]

        },
        'config': {
          'type': 'null',
          'title': 'The Config Schema ',
          'default': null,
          'examples': [
            null
          ]

        }
      }
    }
  };
  public providers = {
    'app': this.app
  };
}
