package main

import (
	"context"
	"errors"
	"fmt"
	"io"
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

func handleConnection(ctx context.Context, conn net.Conn, port serial.Port) (err error) {
	logger := log.MustLogger(ctx)

	logger.Info("Setting TCP no delay")
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetNoDelay(true); err != nil {
			return fmt.Errorf("failed to set TCP no delay: %w", err)
		}
	}

	errCh := make(chan error, 2)

	logger.Info("Copying I/O")
	go func() {
		_, err := io.Copy(conn, port)
		errCh <- err
	}()

	go func() {
		_, err := io.Copy(port, conn)
		errCh <- err
	}()

	err = <-errCh
	err = errors.Join(err, <-errCh)

	err = errors.Join(err, conn.Close())

	logger.Info("Connection closed")

	return
}

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a TCP server connected to a serial port.",
	Long:  "Opens serial port and a TCP server, and pipe communication between both. There's NO security implemented, this can only be used in secure networks at your own risk.",
	Args:  cobra.NoArgs,
	Run: GetRunFn(func(cmd *cobra.Command, args []string) (err error) {

		ctx, logger := log.MustWithAttrs(
			cmd.Context(),
			"port-name", portName,
			"address", address,
			"baud-rate", baudRate,
			"data-bits", dataBits,
			"parity", parity,
			"stop-bits", stopBits,
			"disable-rts", disableRts,
			"disable-dtr", disableDtr,
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

		logger.Info("Opening serial port")
		port, err := serial.Open(portName, mode)
		if err != nil {
			logger.Error("Failed to open serial port", "error", err)
			return err
		}
		defer func() { errors.Join(err, port.Close()) }()

		logger.Info("Listening")
		listener, err := net.Listen("tcp", address)
		if err != nil {
			logger.Error("Failed to listen", "error", err)
			return err
		}
		defer func() { errors.Join(err, listener.Close()) }()

		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Error("Failed to accept connection", "error", err)
				continue
			}
			ctx, logger := log.MustWithGroupAttrs(
				ctx,
				"Connection",
				"LocalAddr", conn.LocalAddr(),
				"RemoteAddr", conn.RemoteAddr(),
			)
			logger.Info("Accepted")

			if err := handleConnection(ctx, conn, port); err != nil {
				logger.Error("Failed to handle connection", "error", err)
			}
		}
	}),
}

func init() {
	ServeCmd.PersistentFlags().StringVarP(&portName, "port-name", "p", portNameDefault, "Port name")
	if err := ServeCmd.MarkFlagRequired("port-name"); err != nil {
		panic(err)
	}
	ServeCmd.PersistentFlags().StringVarP(&address, "address", "a", addressDefault, "TCP address to listen on (host:port)")
	ServeCmd.PersistentFlags().IntVarP(&baudRate, "baud-rate", "b", baudRateDefault, "Serial port baud rate")
	ServeCmd.PersistentFlags().IntVarP(&dataBits, "data-bits", "d", dataBitsDefault, "Serial port data bits (5, 6, 7, or 8)")
	ServeCmd.PersistentFlags().VarP(&parity, "parity", "", "Serial port parity (no, odd, even, mark or space)")
	ServeCmd.PersistentFlags().VarP(&stopBits, "stop-bits", "", "Serial port stop bits (1, 1.5, or 2)")
	ServeCmd.PersistentFlags().BoolVarP(&disableRts, "disable-rts", "", disableRtsDefault, "Serial port RTS (Request To Send)")
	ServeCmd.PersistentFlags().BoolVarP(&disableDtr, "disable-dtr", "", disableDtrDefault, "Serial port DTR (Data Terminal Ready)")

	RootCmd.AddCommand(ServeCmd)
}
