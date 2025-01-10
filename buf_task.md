# Task

## protobuf

Lint
``` sh
buf lint
```

Format
``` sh
buf format -w
```

Detection Breaking
``` sh
buf breaking proto --against '.git#subdir=proto'
```

Solve Deps
``` sh
buf dep update
```

Build
``` sh
buf build
```

Authenticate
``` sh
buf registry login
```

Push
``` sh
buf push --label v1.0.1
```