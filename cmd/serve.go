package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/fornellas/slogxt/log"
	"github.com/kotaira/go-serial"
	"github.com/spf13/cobra"
)

// ParityValue implements pflag.Value for serial.Parity
type ParityValue serial.Parity

func (p *ParityValue) String() string {
	switch serial.Parity(*p) {
	case serial.NoParity:
		return "no"
	case serial.OddParity:
		return "odd"
	case serial.EvenParity:
		return "even"
	case serial.MarkParity:
		return "mark"
	case serial.SpaceParity:
		return "space"
	default:
		return strconv.Itoa(int(*p))
	}
}

func (p *ParityValue) Set(s string) error {
	switch strings.ToLower(s) {
	case "no":
		*p = ParityValue(serial.NoParity)
	case "odd":
		*p = ParityValue(serial.OddParity)
	case "even":
		*p = ParityValue(serial.EvenParity)
	case "mark":
		*p = ParityValue(serial.MarkParity)
	case "space":
		*p = ParityValue(serial.SpaceParity)
	default:
		return fmt.Errorf("invalid parity value: %s", s)
	}
	return nil
}

func (p *ParityValue) Type() string {
	return "parity"
}

// StopBitsValue implements pflag.Value for serial.StopBits
type StopBitsValue serial.StopBits

func (s *StopBitsValue) String() string {
	switch serial.StopBits(*s) {
	case serial.OneStopBit:
		return "1"
	case serial.OnePointFiveStopBits:
		return "1.5"
	case serial.TwoStopBits:
		return "2"
	default:
		return strconv.Itoa(int(*s))
	}
}

func (s *StopBitsValue) Set(str string) error {
	switch strings.ToLower(str) {
	case "1":
		*s = StopBitsValue(serial.OneStopBit)
	case "1.5":
		*s = StopBitsValue(serial.OnePointFiveStopBits)
	case "2":
		*s = StopBitsValue(serial.TwoStopBits)
	default:
		return fmt.Errorf("invalid stop bits value: %s", str)
	}
	return nil
}

func (s *StopBitsValue) Type() string {
	return "bits"
}

var portName string
var portNameDefault = ""

var address string
var addressDefault = "127.0.0.1:9999"

var baudRate int
var baudRateDefault = 115200

var dataBits int
var dataBitsDefault = 8

var parity ParityValue

var stopBits StopBitsValue

var disableRts bool
var disableRtsDefault = false

var disableDtr bool
var disableDtrDefault = false

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
				RTS: !disableRts,
				DTR: !disableDtr,
			},
		}

		port, err := serial.Open(portName, mode)
		if err != nil {
			logger.Error("Failed to open serial port", "error", err)
			return err
		}
		defer func() { errors.Join(err, port.Close()) }()

		listener, err := net.Listen("tcp", address)
		if err != nil {
			logger.Error("Failed to listen", "error", err)
			return err
		}
		defer func() { errors.Join(err, listener.Close()) }()

		logger.Info("Listening on TCP", "address", address, "serial_port", portName)

		return nil
	}),
}

func init() {
	ServeCmd.PersistentFlags().StringVarP(&portName, "port-name", "p", portNameDefault, "Port name")
	ServeCmd.PersistentFlags().StringVarP(&address, "address", "a", addressDefault, "TCP address to listen on (host:port)")
	ServeCmd.PersistentFlags().IntVarP(&baudRate, "baud-rate", "b", baudRateDefault, "Serial port baud rate")
	ServeCmd.PersistentFlags().IntVarP(&dataBits, "data-bits", "d", dataBitsDefault, "Serial port data bits (5, 6, 7, or 8)")
	ServeCmd.PersistentFlags().VarP(&parity, "parity", "", "Serial port parity (no, odd, even, mark or space)")
	ServeCmd.PersistentFlags().VarP(&stopBits, "stop-bits", "", "Serial port stop bits (1, 1.5, or 2)")
	ServeCmd.PersistentFlags().BoolVarP(&disableRts, "disable-rts", "", disableRtsDefault, "Serial port RTS (Request To Send)")
	ServeCmd.PersistentFlags().BoolVarP(&disableDtr, "disable-dtr", "", disableDtrDefault, "Serial port DTR (Data Terminal Ready)")

	RootCmd.AddCommand(ServeCmd)
}
