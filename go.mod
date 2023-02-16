module github.com/islax/microapp

go 1.12

require (
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/containerd/containerd v1.6.18 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/golang-migrate/migrate/v4 v4.15.2
	github.com/golobby/container v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.11.1
	github.com/rs/zerolog v1.18.0
	github.com/satori/go.uuid v1.2.0
	github.com/slok/go-http-metrics v0.9.0
	github.com/spf13/viper v1.14.0
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
	gorm.io/driver/mysql v1.0.4
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.4
)

replace github.com/containerd/containerd v1.6.1 => github.com/containerd/containerd v1.6.12
