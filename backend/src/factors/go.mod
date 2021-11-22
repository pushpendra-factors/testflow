module factors

go 1.14

require (
	cloud.google.com/go/bigquery v1.4.0
	cloud.google.com/go/pubsub v1.1.0
	cloud.google.com/go/storage v1.5.0
	contrib.go.opencensus.io/exporter/stackdriver v0.13.4
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.3.1
	github.com/RichardKnop/logging v0.0.0-20190827224416-1a693bdd4fae
	github.com/RichardKnop/redsync v1.2.0
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/apache/beam v2.21.0+incompatible
	github.com/aws/aws-sdk-go v1.23.20
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/certifi/gocertifi v0.0.0-20200211180108-c7c1fbc02894 // indirect
	github.com/coreos/bbolt v0.0.0-00010101000000-000000000000 // indirect
	github.com/coreos/etcd v3.3.10+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/denisenkom/go-mssqldb v0.11.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/evalphobia/logrus_sentry v0.8.2
	github.com/gamebtc/devicedetector v0.0.0-20191029155033-2b1e13468d79
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/gin-contrib/cors v0.0.0-20170318125340-cf4846e6a636
	github.com/gin-gonic/gin v1.7.0
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/spec v0.19.13 // indirect
	github.com/go-openapi/swag v0.19.12 // indirect
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/golang/snappy v0.0.2-0.20190904063534-ff6b7dc882cf // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/gorilla/rpc v1.1.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.5.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1
	github.com/imdario/mergo v0.3.6
	github.com/jinzhu/copier v0.0.0-20180308034124-7e38e58719c3
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v1.0.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/klauspost/compress v1.10.1 // indirect
	github.com/lib/pq v1.0.0
	github.com/mattn/go-colorable v0.1.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mssola/user_agent v0.5.1
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.17.0 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/oschwald/geoip2-golang v1.2.1
	github.com/oschwald/maxminddb-golang v1.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v0.8.0 // indirect
	github.com/prometheus/common v0.0.0-20180801064454-c7de2306084e // indirect
	github.com/prometheus/procfs v0.0.0-20180725123919-05ee40e3a273 // indirect
	github.com/rs/xid v1.2.1
	github.com/sirupsen/logrus v1.4.2
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/stvp/tempredis v0.0.0-20181119212430-b82af8480203 // indirect
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.3.0
	github.com/swaggo/swag v1.6.7
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.1.0
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	github.com/xdg/stringprep v1.0.1-0.20180714160509-73f8eece6fdc // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.etcd.io/etcd v3.3.10+incompatible
	go.mongodb.org/mongo-driver v1.3.0
	go.opencensus.io v0.22.4
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/exp v0.0.0-20200224162631-6cc2880d07d6 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/api v0.15.1
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/grpc v1.29.1 // Version downgraded from latest. See https://github.com/etcd-io/etcd/issues/11563#issuecomment-580196860.
	gopkg.in/yaml.v2 v2.4.0
	honnef.co/go/tools v0.0.1-2020.1.6 // indirect
)

// Fix for module declares its path as: go.etcd.io/bbolt
//     but was required as: github.com/coreos/bbolt
// https://github.com/etcd-io/etcd/issues/11749#issuecomment-679189808
replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
