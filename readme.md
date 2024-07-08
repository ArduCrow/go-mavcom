# GO MAVCOM

Go Mavcom is a Go library that provides a simple interface to interact with drones that support the Mavlink protocol. It is designed to be simple to use and easy to understand.

## Building

```go build -o bin/appname cmd/main.go```

## Running

With your device connected to a Flight Controller via UART or USB:

```./bin/appname```

If you are connecting over a network/UDP connection, append the ```-t``` flag.

Go Mavcom will then listen on the specified port for a heartbeat.

## Development

If you are writing higher level code that will control an unmanned vehicle, import the vehicle package and create a new Vehicle:

```go
import "gomavlink/pkg/vehicle"

v := vehicle.NewVehicle("/dev/ttyS0", 115200, false)

if err != nil {
    fmt.Println("Error creating Vehicle: ", err)
    return
}
v.Start()
```