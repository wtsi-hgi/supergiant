export class UsersModel {
  user = {
    'model': {
      'username': 'username',
      'password': 'password',
      'role': 'user'
    },
    'schema': {
      'properties': {
        'username': {
          'description': 'User Name',
          'type': 'string'
        },
        'password': {
          'description': 'Password',
          'type': 'string',
        },
        'role': {
          'description': 'User Role',
          'type': 'string',
          'enum': ['user', 'admin'],
          'default': 'user'
        }
      }
    }
  };
}
