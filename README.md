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
- [sql](https://github.com/common-go/signup-sql): requires [gorm](https://github.com/go-gorm/gorm)
- [mongo](https://github.com/common-go/signup-mongo)
- [dynamodb](https://github.com/common-go/signup-dynamodb)
- [firestore](https://github.com/common-go/signup-firestore)
- [elasticsearch](https://github.com/common-go/signup-elasticsearch)
