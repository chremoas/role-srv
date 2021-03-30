module github.com/chremoas/role-srv

go 1.14

require (
	github.com/Masterminds/squirrel v1.5.0
	github.com/chremoas/discord-gateway v1.3.0
	github.com/chremoas/perms-srv v1.3.0
	github.com/chremoas/services-common v1.3.2
	github.com/golang/protobuf v1.3.2
	github.com/jmoiron/sqlx v1.3.1
	github.com/lib/pq v1.10.0
	github.com/micro/go-micro v1.9.1
	github.com/prometheus/common v0.6.0
	github.com/spf13/viper v1.4.0
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80
)

replace github.com/chremoas/role-srv => ../role-srv

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
