module github.com/moov-io/paygate

go 1.13

require (
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-storage-blob-go v0.10.0 // indirect
	github.com/PagerDuty/go-pagerduty v1.3.0
	github.com/PuerkitoBio/goquery v1.5.1 // indirect
	github.com/Shopify/sarama v1.26.4
	github.com/antihax/optional v1.0.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-kit/kit v0.10.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/jaegertracing/jaeger-lib v2.2.0+incompatible
	github.com/jlaffaye/ftp v0.0.0-20200715164256-5d10dd64f695
	github.com/klauspost/compress v1.10.10 // indirect
	github.com/lopezator/migrator v0.3.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/moov-io/ach v1.6.3-0.20210317000541-729e3637e3ba
	github.com/moov-io/base v0.17.0
	github.com/moov-io/customers v0.5.0-rc4.0.20201022164017-d1d2af63aa85
	github.com/moov-io/identity v0.2.7 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/ory/dockertest/v3 v3.6.3
	github.com/ory/mail/v3 v3.0.0
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/pkg/sftp v1.12.0
	github.com/prometheus/client_golang v1.9.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/spf13/afero v1.3.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	gocloud.dev v0.20.0
	gocloud.dev/pubsub/kafkapubsub v0.20.0
	gocloud.dev/secrets/hashivault v0.20.0 // indirect
	goftp.io/server v0.4.0
	golang.org/x/crypto v0.0.0-20201012173705-84dcc777aaee
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/text v0.3.3
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gopkg.in/ini.v1 v1.57.0 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
)

replace goftp.io/server v0.4.0 => github.com/adamdecaf/goftp-server v0.4.0
