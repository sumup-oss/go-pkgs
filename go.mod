module github.com/sumup-oss/go-pkgs

go 1.24.1

replace (
	github.com/sumup-oss/go-pkgs/backoff => ./backoff
	github.com/sumup-oss/go-pkgs/logger => ./logger
)

require (
	github.com/elliotchance/orderedmap v1.2.0
	github.com/mattes/go-expand-tilde v0.0.0-20150330173918-cb884138e64c
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/spf13/cobra v0.0.3
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.10.0
	github.com/sumup-oss/go-pkgs/backoff v0.0.0-00010101000000-000000000000
	github.com/sumup-oss/go-pkgs/logger v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.36.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hashicorp/go-syslog v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/sumup-oss/go-pkgs/errors v1.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
