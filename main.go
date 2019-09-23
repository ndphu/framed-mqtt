package main

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"gocv.io/x/gocv"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	cameraId := os.Getenv("CAMERA_ID")
	if cameraId == "" {
		cameraId = "0"
	}

	broker := os.Getenv("MQTT_BROKER")
	if broker == "" {
		broker = "tcp://35.197.155.112:4443"
	}

	deviceSerial := os.Getenv("DEVICE_SERIAL")

	log.Println("Opening camera with ID", cameraId)

	// open webcam
	webcam, err := gocv.OpenVideoCapture(cameraId)
	if err != nil {
		fmt.Printf("Error opening capture device: %v\n", cameraId)
		return
	}
	defer webcam.Close()

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
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Camera closed: %v\n", cameraId)
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
