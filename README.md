# Rhythm

## Features

* Support for [Docker](https://mesos.apache.org/documentation/latest/docker-containerizer/) and [Mesos](https://mesos.apache.org/documentation/latest/mesos-containerizer/) Containerizers 
* Integration with [HashiCorp Vault](https://www.vaultproject.io/) for secrets management
* Access control list (ACL) backed by [GitLab](https://gitlab.com/)
* [Cron syntax](http://www.nncron.ru/help/EN/working/cron-format.htm)
* Integration with [Sentry](https://sentry.io/) for error tracking

## API

[Documentation](https://mlowicki.github.io/rhythm/api)

## Configuration

Rhythm is configured using file in JSON format. By default config.json from current  directory is used but it can overwritten using `-config` parameter.
There are couple of sections in configuration file:
* api
* storage
* coordinator
* secrets
* mesos
* logging

### API

TODO

### Storage

TODO

### Coordinator

TODO

### Secrets

Secrets backend allow to inject secrets into task via environment variables. Job defines secrets under `secrets` property:
```javascript
"group": "webservices",
"project": "oauth",
"id": "backup",
"secrets": {
    "DB_PASSWORD": "db/password"
}
```

Mesos task will have `DB_PASSWORD` environment variable set to value returned by secrets backend when `"webservices/oauth/db/password"` will be passed. In case of e.g. Vault it'll be interpreted as path to secret from which data under `value` key will retrieved.

Options:
* backend (optional) - `"vault"` or `"none"` (`"none"` by default)
* vault (optional and used only when `backend` is set to `"vault"`)
    * address (required) - Vault server address
    * token (required) - Vault token with read access to secrets under `root`
    * root (optional) - Secret's path prefix (`"secret/rhythm/"` by defualt)
    * timeout (optional) - Client timeout in milliseconds (`0` by default which means no timeout)
    * rootca (optional) - absolute path to custom root certificate used while talking to Vault server
    
Example:
```javascript
"secrets": {
    "backend": "vault",
    "vault": {
        "token": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaaa",
        "address": "https://example.com"
    }
}
```

### Mesos

Options:
* addrs (required) - list of Mesos endpoints
* auth
    * type (optional) - `"none"` or `"basic"` (`"none"` by default)
    * basic (optional and used only when `type` is set to `"basic"`)
        * username (optional)
        * password (optional)
* rootca (optional) - absolute path to custom root certificate used while talking to Mesos
* checkpoint (optional) - controls framework's checkpointing (`false` by default)
* failovertimeout (optional) - number of milliseconds Mesos will wait for the framework to failover before killing all its tasks (7 days used by default)
* hostname (optional) - host for which framework is registered in the Mesos Web UI
* user (optinal) - determine the Unix user that tasks should be launched as
* webuiurl (optional) - framework's Web UI address
* principal (optional) - identifier used while interacting with Mesos
* labels (optional) - dictionary of key-value pairs assigned to framework
* roles (optional) - list of roles framework will subscribe to (`["\*"]` by default)
* logallevents (optional) - print details of all events sent from Mesos (`false` by default)

Example:
```javascript
"mesos": {
    "addrs": ["https://example.com:5050"],
    "principal": "rhythm",
    "roles": ["rhythm"],
    "user": "root",
    "webuiurl": "https://example.com",
    "auth": {
        "type": "basic",
        "basic": {
            "username": "rhythm",
            "password": "secret"
        }
    },
    "labels": {
        "one": "1",
        "two": "2"
    }
}
```

### Logging

Logs are always sent to stderr (`level` defines verbosity) and optional backend to e.g. send certain messages to 3rd party service like Sentry. 

Options:
* level (optional)  - `"debug"`, `"info"`, `"warn"` or `"error"` (`"info"` by default)
* backend (optional) - `"sentry"` or `"none"` (`"none"` by default)
* sentry (optional and used only when `backend` is set to `"sentry"`)

    Logs with level set to warning or error will be sent to Sentry. If logging level is higher than warning then only errors will be sent (in other words `level` defines minium tier which will be by Sentry backend).
    * dsn (required) - Sentry DSN (Data Source Name) passed as string
    * rootca (optional) - absolute path to custom root certificate used while talking to Sentry server
    * tags (optional) - dictionary of custom tags sent with each event

Examples:
```javascript
"logging": {
    "level": "debug",
    "backend": "sentry",
    "sentry": {
        "dsn": "https://key@example.com/123",
        "rootca": "/var/rootca.crt",
        "tags": {
            "one": "1",
            "two": "2"
        }
    }
}
```

```javascript
"logging": {
    "level": "debug"
}
```

There is `-testlogging` option which is used to test events logging. It logs sample error and then program exits. Useful to test backend like Sentry to verify that events are received.
