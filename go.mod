module framed-mqtt

go 1.12

// Consider this option for windows build
// replace gocv.io/x/gocv => D:\\go-path-2\src\\gocv.io\\x\\gocv

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/google/uuid v1.1.1
	gocv.io/x/gocv v0.20.0
	golang.org/x/net v0.7.0 // indirect
)
