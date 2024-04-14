package infinitime

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"tinygo.org/x/bluetooth"
)

const (
	dfuSegmentSize     = 20 // Size of each firmware packet
	dfuPktRecvInterval = 10 // Amount of packets to send before checking for receipt
)

var (
	dfuCmdStart              = []byte{0x01, 0x04}
	dfuCmdRecvInitPkt        = []byte{0x02, 0x00}
	dfuCmdInitPktComplete    = []byte{0x02, 0x01}
	dfuCmdPktReceiptInterval = []byte{0x08}
	dfuCmdRecvFirmware       = []byte{0x03}
	dfuCmdValidate           = []byte{0x04}
	dfuCmdActivateReset      = []byte{0x05}

	dfuResponseStart            = []byte{0x10, 0x01, 0x01}
	dfuResponseInitParams       = []byte{0x10, 0x02, 0x01}
	dfuResponseRecvFwImgSuccess = []byte{0x10, 0x03, 0x01}
	dfuResponseValidate         = []byte{0x10, 0x04, 0x01}
)

// DFUOptions contains options for [UpgradeFirmware]
type DFUOptions struct {
	InitPacket      fs.File
	FirmwareImage   fs.File
	ProgressFunc    func(sent, received, total uint32)
	SegmentSize     int
	ReceiveInterval uint8
}

// UpgradeFirmware upgrades the firmware running on the PineTime.
func (d *Device) UpgradeFirmware(opts DFUOptions) error {
	if opts.SegmentSize <= 0 {
		opts.SegmentSize = dfuSegmentSize
	}

	if opts.ReceiveInterval <= 0 {
		opts.ReceiveInterval = dfuPktRecvInterval
	}

	ctrlPoint, err := d.getChar(dfuCtrlPointChar)
	if err != nil {
		return err
	}

	packet, err := d.getChar(dfuPacketChar)
	if err != nil {
		return err
	}

	d.deviceMtx.Lock()
	defer d.deviceMtx.Unlock()

	d.updating.Store(true)
	defer d.updating.Store(false)

	_, err = ctrlPoint.WriteWithoutResponse(dfuCmdStart)
	if err != nil {
		return err
	}

	fi, err := opts.FirmwareImage.Stat()
	if err != nil {
		return err
	}
	size := uint32(fi.Size())

	sizePacket := make([]byte, 8, 12)
	sizePacket = binary.LittleEndian.AppendUint32(sizePacket, size)
	_, err = packet.WriteWithoutResponse(sizePacket)
	if err != nil {
		return err
	}

	_, err = awaitDFUResponse(ctrlPoint, dfuResponseStart)
	if err != nil {
		return err
	}

	err = writeDFUInitPacket(ctrlPoint, packet, opts.InitPacket)
	if err != nil {
		return err
	}

	err = setRecvInterval(ctrlPoint, opts.ReceiveInterval)
	if err != nil {
		return err
	}

	err = sendFirmware(ctrlPoint, packet, opts, size)
	if err != nil {
		return err
	}

	return finalize(ctrlPoint)
}

func finalize(ctrlPoint *bluetooth.DeviceCharacteristic) error {
	_, err := ctrlPoint.WriteWithoutResponse(dfuCmdValidate)
	if err != nil {
		return err
	}

	_, err = awaitDFUResponse(ctrlPoint, dfuResponseValidate)
	if err != nil {
		return err
	}

	_, _ = ctrlPoint.WriteWithoutResponse(dfuCmdActivateReset)
	return nil
}

func sendFirmware(ctrlPoint, packet *bluetooth.DeviceCharacteristic, opts DFUOptions, totalSize uint32) error {
	_, err := ctrlPoint.WriteWithoutResponse(dfuCmdRecvFirmware)
	if err != nil {
		return err
	}

	var (
		chunksSinceReceipt uint8
		bytesSent          uint32
	)

	chunk := make([]byte, opts.SegmentSize)
	for {
		n, err := opts.FirmwareImage.Read(chunk)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		} else if n == 0 {
			break
		}

		bytesSent += uint32(n)
		_, err = packet.WriteWithoutResponse(chunk[:n])
		if err != nil {
			return err
		}

		if errors.Is(err, io.EOF) {
			break
		}

		chunksSinceReceipt += 1
		if chunksSinceReceipt == opts.ReceiveInterval {
			sizeData, err := awaitDFUResponse(ctrlPoint, []byte{0x11})
			if err != nil {
				return err
			}
			size := binary.LittleEndian.Uint32(sizeData)
			if size != bytesSent {
				return fmt.Errorf("size mismatch: expected %d, got %d", bytesSent, size)
			}
			if opts.ProgressFunc != nil {
				opts.ProgressFunc(bytesSent, size, totalSize)
			}
			chunksSinceReceipt = 0
		}
	}

	return nil
}

func writeDFUInitPacket(ctrlPoint, packet *bluetooth.DeviceCharacteristic, initPkt fs.File) error {
	_, err := ctrlPoint.WriteWithoutResponse(dfuCmdRecvInitPkt)
	if err != nil {
		return err
	}

	initData, err := io.ReadAll(initPkt)
	if err != nil {
		return err
	}

	_, err = packet.WriteWithoutResponse(initData)
	if err != nil {
		return err
	}

	_, err = ctrlPoint.WriteWithoutResponse(dfuCmdInitPktComplete)
	if err != nil {
		return err
	}

	_, err = awaitDFUResponse(ctrlPoint, dfuResponseInitParams)
	return err
}

func setRecvInterval(ctrlPoint *bluetooth.DeviceCharacteristic, interval uint8) error {
	_, err := ctrlPoint.WriteWithoutResponse(append(dfuCmdPktReceiptInterval, interval))
	return err
}

func awaitDFUResponse(ctrlPoint *bluetooth.DeviceCharacteristic, expect []byte) ([]byte, error) {
	respCh := make(chan []byte, 1)
	err := ctrlPoint.EnableNotifications(func(buf []byte) {
		respCh <- buf
	})
	if err != nil {
		return nil, err
	}

	data := <-respCh
	ctrlPoint.EnableNotifications(nil)

	if !bytes.HasPrefix(data, expect) {
		return nil, fmt.Errorf("unexpected dfu response %x (expected %x)", data, expect)
	}

	return bytes.TrimPrefix(data, expect), nil
}
