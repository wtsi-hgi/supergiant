export class ClusterPacketModel {
  packet = {
    'data': {
      'cloud_account_name': '',
      'master_node_size': 'Type 0',
      'kube_master_count': 1,
      'ssh_pub_key': '',
      'name': '',
      'node_sizes': [
        'Type 0',
        'Type 1',
        'Type 2',
        'Type 3',
        'Type 2A'
      ],
      'packet_config': {
        'facility': 'ewr1',
        'project': '',
      }
    },
    'schema': {
      'properties': {
        'cloud_account_name': {
          'description': 'The Supergiant cloud account you created for use with Packet.',
          'type': 'string'
        },
        'master_node_size': {
          'default': 'Type 0',
          'description': 'The size of the server the master will live on.',
          'type': 'string'
        },
        'kube_master_count': {
          'description': 'The number of masters desired--for High Availability.',
          'type': 'number',
          'widget': 'number',
        },
        'ssh_pub_key': {
          'description': 'The public key that will be used to SSH into the kube.',
          'type': 'string',
          'widget': 'textarea',
        },
        'name': {
          'description': 'The desired name of the kube. Max length of 12 characters.',
          'type': 'string',
          'pattern': '^[a-z]([-a-z0-9]*[a-z0-9])?$',
          'maxLength': 12
        },
        'node_sizes': {
          'description': 'The sizes you want to be available to Supergiant when scaling.',
          'items': {
            'type': 'string'
          },
          'type': 'array'
        },
        'packet_config': {
          'properties': {
            'facility': {
              'default': 'ewr1',
              'description': 'The Packet facility (region) the kube will be created in.',
              'type': 'string'
            },
            'project': {
              'description': 'The Packet project the kube will be created in.',
              'type': 'string'
            }
          },
          'type': 'object'
        }
      }
    }
  };
}
