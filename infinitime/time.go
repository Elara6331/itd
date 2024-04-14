package infinitime

import (
	"bytes"
	"encoding/binary"
	"time"
)

// SetTime sets the current time, and then sets the timezone data,
// if the local time characteristic is available.
func (d *Device) SetTime(t time.Time) error {
	c, err := d.getChar(currentTimeChar)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, uint16(t.Year()))
	binary.Write(buf, binary.LittleEndian, uint8(t.Month()))
	binary.Write(buf, binary.LittleEndian, uint8(t.Day()))
	binary.Write(buf, binary.LittleEndian, uint8(t.Hour()))
	binary.Write(buf, binary.LittleEndian, uint8(t.Minute()))
	binary.Write(buf, binary.LittleEndian, uint8(t.Second()))
	binary.Write(buf, binary.LittleEndian, uint8(t.Weekday()))
	binary.Write(buf, binary.LittleEndian, uint8((t.Nanosecond()/1000)/1e6*256))
	binary.Write(buf, binary.LittleEndian, uint8(0b0001))

	_, err = c.WriteWithoutResponse(buf.Bytes())
	if err != nil {
		return err
	}

	ltc, err := d.getChar(localTimeChar)
	if err != nil {
		return nil
	}

	_, offset := t.Zone()
	dst := 0

	// Local time expects two values: the timezone offset and the dst offset, both
	// expressed in quarters of an hour.
	// Timezone offset is to be constant over DST, with dst offset holding the offset != 0
	// when DST is in effect.
	// As there is no standard way in go to get the actual dst offset, we assume it to be 1h
	// when DST is in effect
	if t.IsDST() {
		dst = 3600
		offset -= 3600
	}

	buf.Reset()
	binary.Write(buf, binary.LittleEndian, uint8(offset/3600*4))
	binary.Write(buf, binary.LittleEndian, uint8(dst/3600*4))

	_, err = ltc.WriteWithoutResponse(buf.Bytes())
	return err
}
