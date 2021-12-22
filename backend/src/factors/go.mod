module factors

go 1.14

require (
	cloud.google.com/go/bigquery v1.4.0
	cloud.google.com/go/pubsub v1.1.0
	cloud.google.com/go/storage v1.5.0
	contrib.go.opencensus.io/exporter/stackdriver v0.13.4
	git.apache.org/thrift.git v0.0.0-20180902110319-2566ecd5d999 // indirect
	github.com/360EntSecGroup-Skylar/excelize v1.4.1 // indirect
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.3.1
	github.com/RichardKnop/logging v0.0.0-20190827224416-1a693bdd4fae
	github.com/RichardKnop/redsync v1.2.0
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/apache/beam v2.21.0+incompatible
	github.com/aws/aws-sdk-go v1.34.28
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/campoy/embedmd v0.0.0-20171205015432-c59ce00e0296 // indirect
	github.com/certifi/gocertifi v0.0.0-20200211180108-c7c1fbc02894 // indirect
	github.com/client9/misspell v0.3.4 // indirect
	github.com/coreos/etcd v3.3.10+incompatible
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/dustin/go-broadcast v0.0.0-20171205050544-f664265f5a66 // indirect
	github.com/envoyproxy/go-control-plane v0.9.4 // indirect
	github.com/evalphobia/logrus_sentry v0.8.2
	github.com/gamebtc/devicedetector v0.0.0-20191029155033-2b1e13468d79
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gin-contrib/cors v0.0.0-20170318125340-cf4846e6a636
	github.com/gin-gonic/autotls v0.0.0-20180426091246-be87bd5ef97b // indirect
	github.com/gin-gonic/gin v1.6.3
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/spec v0.19.13 // indirect
	github.com/go-openapi/swag v0.19.12 // indirect
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/gogo/protobuf v1.2.0 // indirect
	github.com/golang/mock v1.4.4 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/golang/snappy v0.0.2-0.20190904063534-ff6b7dc882cf // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/google/martian/v3 v3.0.0 // indirect
	github.com/google/pprof v0.0.0-20200905233945-acf8798be1f7 // indirect
	github.com/google/uuid v1.1.1
	github.com/googleapis/gax-go v1.0.3 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/gorilla/rpc v1.1.0
	github.com/gorilla/securecookie v1.1.1
	github.com/grpc-ecosystem/grpc-gateway v1.5.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1
	github.com/imdario/mergo v0.3.6
	github.com/jessevdk/go-assets v0.0.0-20160921144138-4f4301a06e15 // indirect
	github.com/jinzhu/copier v0.0.0-20180308034124-7e38e58719c3
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v1.0.0
	github.com/jmespath/go-jmespath v0.0.0-20180206201540-c2b33e8439af // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/klauspost/compress v1.10.1 // indirect
	github.com/lib/pq v1.0.0
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/manucorporat/stats v0.0.0-20180402194714-3ba42d56d227 // indirect
	github.com/mattn/go-colorable v0.1.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/mssola/user_agent v0.5.1
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go v0.1.1 // indirect
	github.com/oschwald/geoip2-golang v1.2.1
	github.com/oschwald/maxminddb-golang v1.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v0.8.0 // indirect
	github.com/prometheus/common v0.0.0-20180801064454-c7de2306084e // indirect
	github.com/prometheus/procfs v0.0.0-20180725123919-05ee40e3a273 // indirect
	github.com/richardlehane/mscfb v1.0.3 // indirect
	github.com/richardlehane/msoleps v1.0.1 // indirect
	github.com/rs/xid v1.2.1
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.6.1
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.3.0
	github.com/swaggo/swag v1.6.7
	github.com/thinkerou/favicon v0.1.0 // indirect
	github.com/tmdvs/Go-Emoji-Utils v1.1.0
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.1.0
	github.com/urfave/cli v1.22.5 // indirect
	github.com/urfave/cli/v2 v2.3.0 // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.1-0.20180714160509-73f8eece6fdc // indirect
	github.com/xuri/efp v0.0.0-20200605144744-ba689101faaf // indirect
	go.etcd.io/etcd v3.3.10+incompatible
	go.mongodb.org/mongo-driver v1.5.1
	go.opencensus.io v0.22.4
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/exp v0.0.0-20200224162631-6cc2880d07d6 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/text v0.3.4 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	golang.org/x/tools v0.0.0-20201118030313-598b068a9102 // indirect
	golang.org/x/tools/gopls v0.5.1 // indirect
	google.golang.org/api v0.15.1
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/grpc v1.26.0 // Version downgraded from latest. See https://github.com/etcd-io/etcd/issues/11563#issuecomment-580196860.
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/yaml.v2 v2.3.0
	honnef.co/go/tools v0.0.1-2020.1.6 // indirect
	mvdan.cc/gofumpt v0.0.0-20200927160801-5bfeb2e70dd6 // indirect
	rsc.io/quote/v3 v3.1.0 // indirect
)

// Fix for module declares its path as: go.etcd.io/bbolt
//     but was required as: github.com/coreos/bbolt
// https://github.com/etcd-io/etcd/issues/11749#issuecomment-679189808
replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
