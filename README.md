parse_graphql
=============

Expose a Parse app's schema as a graphql endpoint.

```sh
$ parse_graphql serve -h
Usage:
  parse_graphql [OPTIONS] serve [serve-OPTIONS]

Global options:
  -v, --verbose         Be verbose

Help Options:
  -h, --help            Show this help message

[serve command options]
      -l, --listen=     Listen address (:8080)
      -a, --appID=      Parse Application ID [$PARSE_APPLICATION_ID]
      -m, --masterKey=  Parse Master Key [$PARSE_MASTER_KEY]
      -w, --restApiKey= Parse REST API Key [$PARSE_REST_API_KEY]
```

User signup:

```graphql
mutation signUp { signUp(username: "foobar", password: "bazbar", email: "foo.bar@gmail.com") { objectId, createdAt, sessionToken } }
```

User login:

```graphql
mutation logIn { logIn(username: "foobar", password: "bazbar") { objectId, createdAt, sessionToken } }
```


