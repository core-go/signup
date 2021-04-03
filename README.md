# Sign Up
## Models
- SignUpInfo
- SignUpResult

## Service
- SignUpService

## Installation

Please make sure to initialize a Go module before installing common-go/signup:

```shell
go get -u github.com/common-go/signup
```

Import:

```go
import "github.com/common-go/signup"
```

## Implementations of SignUpRepository
- sql: SqlSignUpRepository
- [mongo](https://github.com/common-go/signup-mongo)
- [dynamodb](https://github.com/common-go/signup-dynamodb)
- [firestore](https://github.com/common-go/signup-firestore)
- [elasticsearch](https://github.com/common-go/signup-elasticsearch)
