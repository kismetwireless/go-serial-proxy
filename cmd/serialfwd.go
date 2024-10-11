package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"go.bug.st/serial"
)

var (
	listMode   bool
	helpMode   bool
	serialPort string
	tcpPort    string
	speed      int
)

func forward(conn net.Conn) {
	mode := &serial.Mode{
		BaudRate: speed,
	}

	port, err := serial.Open(serialPort, mode)
	if err != nil {
		log.Fatal(err)
		conn.Write([]byte("ERROR: " + err.Error()))
		return
	}

	go func() {
		defer port.Close()
		defer conn.Close()
		io.Copy(port, conn)
	}()
	go func() {
		defer port.Close()
		defer conn.Close()
		io.Copy(conn, port)
	}()
}

func main() {
	flag.BoolVar(&listMode, "list", false, "List serial ports")
	flag.BoolVar(&helpMode, "help", false, "Help")
	flag.StringVar(&serialPort, "device", "/dev/cu.usbserial-foo", "Path to serial port (typically /dev/cu.usbserial-SOMETHING)")
	flag.IntVar(&speed, "speed", 9600, "Serial port speed")
	flag.StringVar(&tcpPort, "port", "8888", "TCP port")
	flag.Parse()

	if helpMode {
		fmt.Printf("SERIALFWD\n")
		fmt.Printf("Forward a serial port device to a TCP socket for use with TCP based tools like MuffinTerm\n")
		fmt.Printf("\n")
		fmt.Printf("Usage:  serialfwd --device=/path/to-device [--speed=serial-speed] [--tcpport=tcp-port-number]\n")
		fmt.Printf("  --list            List all serial port devices on the system\n")
		fmt.Printf("  --help            This.\n")
		fmt.Printf("  --device=[path]   Path to serial port, such as /dev/cu.usbserial-32310\n")
		fmt.Printf("  --speed=[speed]   Serial port speed (default 9600)\n")
		fmt.Printf("  --tcpport=[port]     TCP Port number (defaults to 8888) for the serial port mirror\n")
		return
	}

	if listMode {
		ports, err := serial.GetPortsList()

		if err != nil {
			log.Fatal(err)
		}

		if len(ports) == 0 {
			log.Fatal("No serial ports found!")
		}

		for _, port := range ports {
			if port[:8] == "/dev/cu." {
				fmt.Printf("Found serial device: %v\n", port)
			}
		}

		return
	}

	localAddr := ":" + tcpPort
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatalf("Failed to setup listener: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("ERROR: failed to accept listener: %v", err)
		}

		log.Printf("Accepted connection from %v\n", conn.RemoteAddr().String())
		go forward(conn)
	}
}
