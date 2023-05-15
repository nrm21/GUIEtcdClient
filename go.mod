module _nate/EtcdChat

go 1.17

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/coreos/etcd v3.3.27+incompatible
	github.com/etcd-io/etcd v3.3.27+incompatible
	github.com/lxn/walk v0.0.0-20201209144500-98655d01b2f1
	github.com/nrm21/support v0.0.0-20230515013121-48b783201b16
	go.etcd.io/etcd v3.3.27+incompatible
	golang.org/x/sys v0.8.0
	google.golang.org/grpc v1.55.0
	gopkg.in/yaml.v2 v2.2.8
	sigs.k8s.io/yaml v1.2.0
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20230327231512-ba87abf18a23 // indirect
	github.com/envoyproxy/go-control-plane v0.11.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/lxn/win v0.0.0-20201111105847-2a20daff6a55 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
)
