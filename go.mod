module github.com/stitchfix/flotilla-os

go 1.12

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Bowery/prompt v0.0.0-20190916142128-fa8279994f75 // indirect
	github.com/Microsoft/go-winio v0.4.13-0.20190625174015-d2ef9cfdac5d
	github.com/Microsoft/hcsshim v0.0.0-20190627211051-c6f98528dede
	github.com/aws/aws-sdk-go v1.15.11
	github.com/beorn7/perks v1.0.0
	github.com/containerd/continuity v0.0.0-20190426062206-aaeac12a7ffc
	github.com/davecgh/go-spew v1.1.1
	github.com/dchest/safefile v0.0.0-20151022103144-855e8d98f185 // indirect
	github.com/docker/cli v0.0.0-20190702184337-39e22d9db677
	github.com/docker/distribution v0.0.0-20190628181051-be07be99045e
	github.com/docker/docker v0.0.0-20190702170247-a43a2ed74654
	github.com/docker/docker-credential-helpers v0.0.0-20190620125321-680ca48e6d4a
	github.com/docker/go-connections v0.0.0-20190612165340-fd1b1942c4d5
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82
	github.com/docker/go-units v0.4.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-kit/kit v0.9.0
	github.com/go-logfmt/logfmt v0.4.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/google/gofuzz v1.0.0
	github.com/google/shlex v0.0.0-20181106134648-c34317bd91bf // indirect
	github.com/googleapis/gnostic v0.0.0-20191023004841-dde5565d9866
	github.com/gorilla/mux v0.0.0-20190701202633-d83b6ffe499a
	github.com/hashicorp/errwrap v1.0.0
	github.com/hashicorp/hcl v1.0.0
	github.com/imdario/mergo v0.3.8
	github.com/jmespath/go-jmespath v0.0.0-20180206201540-c2b33e8439af
	github.com/jmoiron/sqlx v0.0.0-20190426154859-38398a30ed85
	github.com/json-iterator/go v1.1.7
	github.com/kardianos/govendor v1.0.9 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2
	github.com/kr/logfmt v0.0.0-20140226030751-b84e30acd515
	github.com/lib/pq v1.0.0
	github.com/magiconair/properties v1.8.1
	github.com/mattn/go-shellwords v0.0.0-20190425161501-2444a32a19f4
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mitchellh/mapstructure v1.1.2
	github.com/moby/moby v0.0.0-20190702170247-a43a2ed74654
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.1
	github.com/morikuni/aec v1.0.0
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/opencontainers/go-digest v0.0.0-20190228220655-ac19fd6e7483
	github.com/opencontainers/image-spec v1.0.0
	github.com/opencontainers/runc v0.0.0-20190626165814-6cccc1760d57
	github.com/patrickmn/go-cache v0.0.0-20180815053127-5633e0862627
	github.com/pelletier/go-toml v1.2.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/prometheus/common v0.6.0
	github.com/prometheus/procfs v0.0.2
	github.com/rs/cors v0.0.0-20190613161432-33ffc0734c60
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/afero v1.2.2
	github.com/spf13/cast v1.3.0
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v0.0.0-20190614151712-3349bd9cc288
	github.com/subosito/gotenv v1.1.1
	golang.org/x/crypto v0.0.0-20191108234033-bd318be0434a
	golang.org/x/net v0.0.0-20191108225301-c7154b74f18f
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456
	golang.org/x/text v0.3.2
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135 // indirect
	google.golang.org/genproto v0.0.0-20190701230453-710ae3a149df
	google.golang.org/grpc v1.21.0
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	gopkg.in/yaml.v2 v2.2.4
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
	k8s.io/api v0.0.0-20191108065827-59e77acf588f
	k8s.io/apimachinery v0.0.0-20191108065633-c18f71bf2947
	k8s.io/client-go v0.0.0-20191108070106-f8f007fd456c
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20191030222137-2b95a09bc58d
	sigs.k8s.io/yaml v1.1.0
)
