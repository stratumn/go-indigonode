module github.com/stratumn/go-node

go 1.12

require (
	cloud.google.com/go v0.32.0 // indirect
	code.cloudfoundry.org/bytefmt v0.0.0-20180906201452-2aa6f33b730c
	contrib.go.opencensus.io/exporter/stackdriver v0.7.0
	git.apache.org/thrift.git v0.0.0-20181106172052-f7d43ce0aa58 // indirect
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/aws/aws-sdk-go v1.15.72 // indirect
	github.com/c-bata/go-prompt v0.2.3
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/coreos/etcd v3.3.10+incompatible
	github.com/gibson042/canonicaljson-go v1.0.3
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.2.1
	github.com/golang/mock v1.1.1
	github.com/golang/protobuf v1.3.0
	github.com/hpcloud/tail v1.0.0
	github.com/huin/goupnp v1.0.0 // indirect
	github.com/improbable-eng/grpc-web v0.0.0-20181031170435-f683dbb3b587
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/ipfs/go-cid v0.0.1
	github.com/ipfs/go-ds-leveldb v0.0.1
	github.com/ipfs/go-log v0.0.1
	github.com/jackpal/gateway v1.0.5 // indirect
	github.com/jbenet/go-base58 v0.0.0-20150317085156-6237cf65f3a6
	github.com/jbenet/goprocess v0.0.0-20160826012719-b497e2f366b8
	github.com/jhump/protoreflect v0.0.0-20181105161220-9cb68d9a8322
	github.com/libp2p/go-libp2p v0.0.2
	github.com/libp2p/go-libp2p-blankhost v0.0.1
	github.com/libp2p/go-libp2p-circuit v0.0.1
	github.com/libp2p/go-libp2p-connmgr v0.0.1
	github.com/libp2p/go-libp2p-crypto v0.0.1
	github.com/libp2p/go-libp2p-host v0.0.1
	github.com/libp2p/go-libp2p-interface-connmgr v0.0.1
	github.com/libp2p/go-libp2p-interface-pnet v0.0.1
	github.com/libp2p/go-libp2p-kad-dht v0.0.4
	github.com/libp2p/go-libp2p-kbucket v0.1.0 // indirect
	github.com/libp2p/go-libp2p-metrics v0.0.1
	github.com/libp2p/go-libp2p-net v0.0.1
	github.com/libp2p/go-libp2p-peer v0.0.1
	github.com/libp2p/go-libp2p-peerstore v0.0.1
	github.com/libp2p/go-libp2p-protocol v0.0.1
	github.com/libp2p/go-libp2p-pubsub v0.0.1
	github.com/libp2p/go-libp2p-secio v0.0.1
	github.com/libp2p/go-libp2p-swarm v0.0.1
	github.com/libp2p/go-libp2p-transport v0.0.4
	github.com/libp2p/go-libp2p-transport-upgrader v0.0.1
	github.com/libp2p/go-maddr-filter v0.0.1
	github.com/libp2p/go-stream-muxer v0.0.1
	github.com/libp2p/go-tcp-transport v0.0.1
	github.com/libp2p/go-testutil v0.0.1
	github.com/mattn/go-runewidth v0.0.0-20181025052659-b20a3daf6a39 // indirect
	github.com/mattn/go-tty v0.0.0-20180907095812-13ff1204f104 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/multiformats/go-multiaddr v0.0.2
	github.com/multiformats/go-multiaddr-dns v0.0.2
	github.com/multiformats/go-multiaddr-net v0.0.1
	github.com/multiformats/go-multicodec v0.1.6
	github.com/multiformats/go-multihash v0.0.1
	github.com/multiformats/go-multistream v0.0.1
	github.com/mwitkow/go-conntrack v0.0.0-20161129095857-cc309e4a2223 // indirect
	github.com/pelletier/go-toml v1.2.0
	github.com/pkg/errors v0.8.1
	github.com/pkg/term v0.0.0-20181103235908-93e6c9149309 // indirect
	github.com/prometheus/client_golang v0.9.1 // indirect
	github.com/prometheus/common v0.0.0-20181109100915-0b1957f9d949 // indirect
	github.com/prometheus/procfs v0.0.0-20181005140218-185b4288413d // indirect
	github.com/rs/cors v1.6.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.2.1
	github.com/stratumn/merkle v0.0.0-20181025134551-a15850ce9757
	github.com/stretchr/testify v1.3.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/takama/daemon v0.0.0-20180403113744-aa76b0035d12
	github.com/whyrusleeping/go-smux-multistream v2.0.2+incompatible
	github.com/whyrusleeping/go-smux-yamux v2.0.9+incompatible
	github.com/whyrusleeping/multiaddr-filter v0.0.0-20160516205228-e903e4adabd7
	go.opencensus.io v0.18.0
	golang.org/x/oauth2 v0.0.0-20181106182150-f42d05182288 // indirect
	golang.org/x/sync v0.0.0-20181108010431-42b317875d0f // indirect
	google.golang.org/api v0.0.0-20181108001712-cfbc873f6b93 // indirect
	google.golang.org/appengine v1.3.0 // indirect
	google.golang.org/genproto v0.0.0-20181107211654-5fc9ac540362 // indirect
	google.golang.org/grpc v1.16.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0-20170531160350-a96e63847dc3
)
