package main

import (
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"gocv.io/x/gocv"
	"log"
	"os"
	"os/signal"
	"time"
)

type Config struct {
	MQTTBroker   string `json:"mqttBroker"`
	MQTTUsername string `json:"mqttUsername"`
	MQTTPassword string `json:"mqttPassword"`
	CameraId     string `json:"cameraId"`
	DeviceSerial string `json:"deviceSerial"`
}

var config Config

func init() {
	if file, err := os.Open("config.json"); err != nil {
		log.Fatal("Fail to load config:", err)
	} else {
		if err := json.NewDecoder(file).Decode(&config); err != nil {
			log.Fatal("Fail to load config:", err)
		} else {
			validateConfig()
			log.Println("Config loaded successfully")
		}
	}
}

func validateConfig() {
	if config.MQTTBroker == "" {
		log.Panic("Missing MQTTBroker")
	}
	if config.CameraId == "" {
		log.Panic("Missing CameraId")
	}
	if config.DeviceSerial == "" {
		log.Panic("Missing DeviceSerial")
	}
}

func main() {
	broker := config.MQTTBroker
	deviceSerial := config.DeviceSerial

	log.Println("Opening camera with ID", config.CameraId)
	cam, err := gocv.OpenVideoCapture(config.CameraId)
	if err != nil {
		log.Fatalf("Error opening capture device: %s by error: %v\n", config.CameraId, err)
	}
	defer cam.Close()

	log.Println("Camera opened successfully.")

	clientId := "framed-" + uuid.New().String()
	log.Println("Connecting to MQTT", broker, "using client ID:", clientId)

	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientId)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topic := "/3ml/device/" + deviceSerial + "/framed/out"

	log.Println("Publishing frame to topic", topic)

	img := gocv.NewMat()
	defer img.Close()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	for {
		if ok := cam.Read(&img); !ok {
			log.Fatalf("Camera closed: %s\n", config.CameraId)
			return
		}
		if img.Empty() {
			continue
		}

		buf, _ := gocv.IMEncode(".jpg", img)
		c.Publish(topic, 0, false, buf).Wait()
		select {
		case <-signalChan:
			log.Println("Interrupt signal received. Exiting...")
			return
		default:
			continue
		}
	}
}
