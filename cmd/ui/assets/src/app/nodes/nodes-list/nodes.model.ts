export class NodesModel {
  node = {
    'model': {
      'kube_name': '',
      'size': '',
    },
    'schema': {
      'properties': {
        'kube_name': {
          'type': 'string'
        },
        'size': {
          'type': 'string',
          'widget': 'select',
          'enum': [],
          'default': '',
        }
      }
    }
  };
  public providers = {
    'node': this.node
  };
}
