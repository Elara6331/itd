package fsproto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"

	"tinygo.org/x/bluetooth"
)

type FSReqOpcode uint8

const (
	ReadFileHeaderOpcode  FSReqOpcode = 0x10
	ReadFileOpcode        FSReqOpcode = 0x12
	WriteFileHeaderOpcode FSReqOpcode = 0x20
	WriteFileOpcode       FSReqOpcode = 0x22
	DeleteFileOpcode      FSReqOpcode = 0x30
	MakeDirectoryOpcode   FSReqOpcode = 0x40
	ListDirectoryOpcode   FSReqOpcode = 0x50
	MoveFileOpcode        FSReqOpcode = 0x60
)

type FSRespOpcode uint8

const (
	ReadFileResp      FSRespOpcode = 0x11
	WriteFileResp     FSRespOpcode = 0x21
	DeleteFileResp    FSRespOpcode = 0x31
	MakeDirectoryResp FSRespOpcode = 0x41
	ListDirectoryResp FSRespOpcode = 0x51
	MoveFileResp      FSRespOpcode = 0x61
)

type ReadFileHeaderRequest struct {
	Padding byte
	PathLen uint16
	Offset  uint32
	ReadLen uint32
	Path    string
}

type ReadFileRequest struct {
	Status  uint8
	Padding [2]byte
	Offset  uint32
	ReadLen uint32
}

type ReadFileResponse struct {
	Status   int8
	Padding  [2]byte
	Offset   uint32
	FileSize uint32
	ChunkLen uint32
	Data     []byte
}

type WriteFileHeaderRequest struct {
	Padding  byte
	PathLen  uint16
	Offset   uint32
	ModTime  uint64
	FileSize uint32
	Path     string
}

type WriteFileRequest struct {
	Status   uint8
	Padding  [2]byte
	Offset   uint32
	ChunkLen uint32
	Data     []byte
}

type WriteFileResponse struct {
	Status    int8
	Padding   [2]byte
	Offset    uint32
	ModTime   uint64
	FreeSpace uint32
}

type DeleteFileRequest struct {
	Padding byte
	PathLen uint16
	Path    string
}

type DeleteFileResponse struct {
	Status int8
}

type MkdirRequest struct {
	Padding   byte
	PathLen   uint16
	Padding2  [4]byte
	Timestamp uint64
	Path      string
}

type MkdirResponse struct {
	Status  int8
	Padding [6]byte
	ModTime uint64
}

type ListDirRequest struct {
	Padding byte
	PathLen uint16
	Path    string
}

type ListDirResponse struct {
	Status       int8
	PathLen      uint16
	EntryNum     uint32
	TotalEntries uint32
	Flags        uint32
	ModTime      uint64
	FileSize     uint32
	Path         []byte
}

type MoveFileRequest struct {
	Padding    byte
	OldPathLen uint16
	NewPathLen uint16
	OldPath    string
	Padding2   byte
	NewPath    string
}

type MoveFileResponse struct {
	Status int8
}

func WriteRequest(char *bluetooth.DeviceCharacteristic, opcode FSReqOpcode, req any) error {
	buf := &bytes.Buffer{}
	buf.WriteByte(byte(opcode))

	rv := reflect.ValueOf(req)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	for i := 0; i < rv.NumField(); i++ {
		switch field := rv.Field(i); field.Kind() {
		case reflect.String:
			io.WriteString(buf, field.String())
		case reflect.Slice:
			if field.Type().Elem().Kind() == reflect.Uint8 {
				buf.Write(field.Bytes())
			}
		default:
			binary.Write(buf, binary.LittleEndian, field.Interface())
		}
	}

	_, err := char.WriteWithoutResponse(buf.Bytes())
	return err
}

func ReadResponse(b []byte, expect FSRespOpcode, out interface{}) error {
	if len(b) == 0 {
		return errors.New("empty response packet")
	}
	if opcode := FSRespOpcode(b[0]); opcode != expect {
		return fmt.Errorf("unexpected response opcode: expected %x, got %x", expect, opcode)
	}

	r := bytes.NewReader(b[1:])

	ot := reflect.TypeOf(out)
	if ot.Kind() != reflect.Ptr || ot.Elem().Kind() != reflect.Struct {
		return errors.New("out parameter must be a pointer to a struct")
	}

	ov := reflect.ValueOf(out).Elem()
	for i := 0; i < ot.Elem().NumField(); i++ {
		field := ot.Elem().Field(i)
		fieldValue := ov.Field(i)

		// If the last field is a byte slice, just read the remaining data into it and return.
		if i == ot.Elem().NumField()-1 {
			if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Uint8 {
				data, err := io.ReadAll(r)
				if err != nil {
					return err
				}
				fieldValue.SetBytes(data)
				return nil
			}
		}

		if err := binary.Read(r, binary.LittleEndian, fieldValue.Addr().Interface()); err != nil {
			return err
		}
	}

	if statusField := ov.FieldByName("Status"); !statusField.IsZero() {
		code := statusField.Interface().(int8)
		if code != 0x01 {
			return Error{code}
		}
	}

	return nil
}
