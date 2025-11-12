package main

import (
	"net"

	"github.com/fornellas/slogxt/log"
	"github.com/kotaira/go-serial"
	"github.com/spf13/cobra"
)

var portName string
var portNameDefault = ""

var address string
var addressDefault = "127.0.0.1:9999"

var baudRate int
var baudRateDefault = 115200

var dataBits int
var dataBitsDefault = 8

var parity int
var parityDefault = 0 // serial.NoParity

var stopBits int
var stopBitsDefault = 0 // serial.OneStopBit

var rts bool
var rtsDefault = false

var dtr bool
var dtrDefault = false

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a TCP server connected to a serial port.",
	Args:  cobra.ExactArgs(1),
	Run: GetRunFn(func(cmd *cobra.Command, args []string) (err error) {

		ctx, logger := log.MustWithAttrs(
			cmd.Context(),
		)
		cmd.SetContext(ctx)
		logger.Info("Running")

		mode := &serial.Mode{
			BaudRate: baudRate,
			DataBits: dataBits,
			Parity:   serial.Parity(parity),
			StopBits: serial.StopBits(stopBits),
			InitialStatusBits: &serial.ModemOutputBits{
				RTS: rts,
				DTR: dtr,
			},
		}

		port, err := serial.Open(portName, mode)
		if err != nil {
			logger.Error("Failed to open serial port", "error", err)
			return err
		}
		defer port.Close()

		listener, err := net.Listen("tcp", address)
		if err != nil {
			logger.Error("Failed to listen", "error", err)
			return err
		}
		defer listener.Close()

		logger.Info("Listening on TCP", "address", address, "serial_port", portName)

		return nil
	}),
}

func init() {
	ServeCmd.PersistentFlags().StringVarP(&portName, "port-name", "p", portNameDefault, "Port name")
	ServeCmd.PersistentFlags().StringVarP(&address, "address", "a", addressDefault, "TCP address to listen on (host:port)")
	ServeCmd.PersistentFlags().IntVarP(&baudRate, "baud-rate", "b", baudRateDefault, "Serial port baud rate")
	ServeCmd.PersistentFlags().IntVarP(&dataBits, "data-bits", "d", dataBitsDefault, "Serial port data bits (5, 6, 7, or 8)")
	ServeCmd.PersistentFlags().IntVarP(&parity, "parity", "", parityDefault, "Serial port parity (0=NoParity, 1=OddParity, 2=EvenParity, 3=MarkParity, 4=SpaceParity)")
	ServeCmd.PersistentFlags().IntVarP(&stopBits, "stop-bits", "", stopBitsDefault, "Serial port stop bits (0=OneStopBit, 1=OnePointFiveStopBits, 2=TwoStopBits)")
	ServeCmd.PersistentFlags().BoolVarP(&rts, "rts", "", rtsDefault, "Serial port RTS (Request To Send) initial status")
	ServeCmd.PersistentFlags().BoolVarP(&dtr, "dtr", "", dtrDefault, "Serial port DTR (Data Terminal Ready) initial status")

	RootCmd.AddCommand(ServeCmd)
}
