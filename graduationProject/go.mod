module graduation

go 1.16

require (
	github.com/go-redis/redis/v8 v8.11.4
	github.com/jinzhu/gorm v1.9.16
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
)

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

replace google.golang.org/grpc v1.32.0 => google.golang.org/grpc v1.26.0
