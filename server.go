package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

func main() {

  mac := []byte{0xb8, 0x27, 0xeb, 0x21, 0x16, 0x3d}
  // steps little endian
  // 01 （步数）10 27 00（1万步） 0x002710 = 10000
  // http://iot.weixin.qq.com/wiki/new/index.html?page=4-3
  //steps := []byte{0x01, 0xe7, 0x22, 0x00}
  steps := []byte{0x01, byte(rand.Intn(255)), byte(rand.Intn(18))+0x1c, 0x00}

	const (
		flagLimitedDiscoverable = 0x01 // LE Limited Discoverable Mode
		flagGeneralDiscoverable = 0x02 // LE General Discoverable Mode
		flagLEOnly              = 0x04 // BR/EDR Not Supported. Bit 37 of LMP Feature Mask Definitions (Page 0)
		flagBothController      = 0x08 // Simultaneous LE and BR/EDR to Same Device Capable (Controller).
		flagBothHost            = 0x10 // Simultaneous LE and BR/EDR to Same Device Capable (Host).
	)

	d, err := gatt.NewDevice(option.DefaultServerOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s", err)
	}

	// Register optional handlers.
	d.Handle(
		gatt.CentralConnected(func(c gatt.Central) { fmt.Println("Connect: ", c.ID()) }),
		gatt.CentralDisconnected(func(c gatt.Central) { fmt.Println("Disconnect: ", c.ID()) }),
	)

	// A mandatory handler for monitoring device state.
	onStateChanged := func(d gatt.Device, s gatt.State) {
		fmt.Printf("State: %s\n", s)
		switch s {
		case gatt.StatePoweredOn:

			s0 := gatt.NewService(gatt.UUID16(0xFEE7))

			c0 := s0.AddCharacteristic(gatt.UUID16(0xFEA1))
			c0.HandleReadFunc(
				func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
					log.Println("Read: 0xFEA1")
					rsp.Write(steps)
				})
			c0.HandleNotifyFunc(
				func(r gatt.Request, n gatt.Notifier) {
					go func() {
						n.Write(steps)
						log.Printf("Indicate 0xFEA2")
					}()
				})

			c1 := s0.AddCharacteristic(gatt.UUID16(0xFEA2))
			c1.HandleReadFunc(
				func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
					log.Println("Read: 0xFEA2")
					rsp.Write(steps)
				})
			c1.HandleNotifyFunc(
				func(r gatt.Request, n gatt.Notifier) {
					go func() {
						n.Write(steps)
						log.Printf("Indicate 0xFEA2")
					}()
				})

			c1.HandleWriteFunc(
				func(r gatt.Request, data []byte) (status byte) {
					log.Println("Wrote 0xFEA2:", string(data))
					return gatt.StatusSuccess
				})

			c2 := s0.AddCharacteristic(gatt.UUID16(0xFEC9))
			c2.HandleReadFunc(
				func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
					log.Println("Read: 0xFEC9")
					rsp.Write(mac)
				})

			d.AddService(s0)
			// Advertise device name and service's UUIDs.
			a := &gatt.AdvPacket{}
			a.AppendFlags(flagGeneralDiscoverable | flagLEOnly)
			a.AppendUUIDFit([]gatt.UUID{s0.UUID()})
			a.AppendName("salt")
			// company id and data, MAC Address
			// https://www.bluetooth.com/specifications/assigned-numbers/company-identifiers
			a.AppendManufacturerData(0x2333, mac)

			d.Advertise(a)

		default:
		}
	}

	d.Init(onStateChanged)
	select {}
}
