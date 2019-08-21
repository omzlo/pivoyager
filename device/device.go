package device

import (
	"errors"
	"fmt"
	"github.com/omzlo/pivoyager/i2c"
	"strings"
	"time"
)

type DeviceStatus byte

/*  TODO: DeviceStatus and ConfiguratioByte should probably be factored together */

var (
	stateBits = [8]string{"pg", "stat1", "stat2", "5v", "inits", "!B5!", "alarm", "button"}
	batState  = [8]string{"n/a", "fault", "err", "charge complete", "low battery", "charging", "discharging", "no battery"}
)

func (s DeviceStatus) BatteryStateString() string {
	return batState[s&7]
}

func FromBCD(bcd byte) int {
	return int((bcd>>4)*10 + (bcd & 0xF))
}
func ToBCD(i int) byte {
	return byte((i/10)<<4) + byte(i%10)
}

func (s DeviceStatus) ToStrings() []string {
	var i byte
	var res []string

	for i = 0; i < 8; i++ {
		if (s & (1 << i)) != 0 {
			res = append(res, stateBits[i])
		}
	}
	return res
}

func (s *DeviceStatus) FromStrings(ss []string) error {
	*s = 0
	for _, v := range ss {
		candidate := strings.ToLower(v)
		found := false
		for i, k := range stateBits {
			if candidate == k {
				*s |= DeviceStatus(1) << uint(i)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Invalid status bit: %s", v)
		}
	}
	return nil
}

func (s DeviceStatus) String() string {
	return strings.Join(s.ToStrings(), " ")
}

/*******/

const (
	REG_MODE      = 0
	REG_STAT      = 1
	REG_CONF      = 2
	REG_PROG      = 3

	REG_TIME      = 4
	REG_DATE      = 8

	REG_SET_TIME  = 12
	REG_SET_DATE  = 16

	REG_WATCH     = 20
	REG_WAKE      = 22

	REG_ALARM       = 24
	REG_BOOT        = 28
    REG_FW_VERSION  = 30

	REG_VBAT      = 32
	REG_VREF      = 34
	REG_VREF_CAL  = 36
	REG_LBO_TIMER = 38
	//total size = 40
)

const (
	DEVICE_ADDRESS = 0x65
)

const (
	CONF_I2C_WD       = 0x01
	CONF_PIN_WD       = 0x02
	CONF_WAKE_AFTER   = 0x04
	CONF_WAKE_ALARM   = 0x08
	CONF_WAKE_POWER   = 0x10
	CONF_WAKE_BUTTON  = 0x20
    CONF_LBO_SHUTDOWN = 0x80
)

const (
    PROG_BOOTLOADER   = 0x01
	PROG_CLEAR_ALARM  = 0x10
	PROG_CLEAR_BUTTON = 0x20
	PROG_CALENDAR     = 0x40
	PROG_ALARM        = 0x80
)

var conf_strings = []string{"i2c-watchdog", "gpio-watchdog", "timer-wakeup", "alarm-wakeup", "power-wakeup", "button-wakeup", "undefined", "low-battery-shutdown"}

type ConfigurationByte byte

func (c ConfigurationByte) ToStrings() []string {
	var i byte
	var res []string

	for i = 0; i < 8; i++ {
		if (c & (1 << i)) != 0 {
			res = append(res, conf_strings[i])
		}
	}
	return res
}

func (c *ConfigurationByte) FromStrings(s []string) error {
	*c = 0
	for _, v := range s {
		candidate := strings.ToLower(v)
		found := false
		for i, k := range conf_strings {
			if candidate == k {
				*c |= ConfigurationByte(1) << uint(i)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Invalid configuration option: %s", v)
		}
	}
	return nil
}

func (c ConfigurationByte) Invert() ConfigurationByte {
	return (^c) & 0x3F
}

func (c ConfigurationByte) String() string {
	return strings.Join(c.ToStrings(), " ")
}

type Device struct {
	i2c.Bus
	address byte
}

var (
	ModeError error = errors.New("Device in incorrect mode")
)

func Open(bootloader bool) (*Device, error) {
	bus := i2c.OpenBus(1)
	r, err := bus.ReadByte(DEVICE_ADDRESS, REG_MODE)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to i2c device: %s", err)
	}

	if r != 'N' && r != 'B' {
		return nil, fmt.Errorf("Unrecognized signature byte 0x%02x. i2c device does not seem to be a pivoyager", r)
	}
	if (bootloader && r != 'B') || (!bootloader && r != 'N') {
		return nil, ModeError
	}
	return &Device{bus, DEVICE_ADDRESS}, nil
}

func (dev *Device) FirmwareVersion() (string, error) {
    var buf [2]byte

    if err := dev.ReadBytes(dev.address, REG_FW_VERSION, buf[:]); err != nil {
        return "", err
    }
    return fmt.Sprintf("%x.%02x", buf[1], buf[0]), nil
}

func (dev *Device) Time() (time.Time, error) {
	var buf [8]byte

	if err := dev.ReadBytes(dev.address, REG_TIME, buf[:]); err != nil {
		return time.Unix(0, 0), err
	}
	//fmt.Printf("Receiving %s\n", hex.EncodeToString(buf[:]))
	return time.Date(2000+FromBCD(buf[6]), time.Month(FromBCD(buf[5]&0x1F)), FromBCD(buf[4]), FromBCD(buf[2]), FromBCD(buf[1]), FromBCD(buf[0]), 0, time.UTC), nil
}

func (dev *Device) SetTime(tm time.Time) error {
	var buf [8]byte
	buf[0] = ToBCD(tm.Second())
	buf[1] = ToBCD(tm.Minute())
	buf[2] = ToBCD(tm.Hour())
	buf[3] = 0
	buf[4] = ToBCD(tm.Day())
	weekday := byte(tm.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	buf[5] = ToBCD(int(tm.Month())) | (weekday << 5)
	buf[6] = ToBCD(tm.Year() % 100)
	buf[7] = 0
	//fmt.Printf("Sending %s\n", hex.EncodeToString(buf[:]))
	if err := dev.WriteBytes(dev.address, REG_SET_TIME, buf[:]); err != nil {
		return err
	}
	return dev.Program(PROG_CALENDAR)
}

func (dev *Device) Status() (DeviceStatus, error) {
	r, err := dev.ReadByte(dev.address, REG_STAT)
	if err != nil {
		return 0, err
	}
	return DeviceStatus(r), nil
}

func (dev *Device) Program(b byte) error {
	return dev.WriteByte(dev.address, REG_PROG, b)
}

func (dev *Device) Voltage() (float32, float32, error) {
	var buf [6]byte
	var vbat, vref, vcal uint16

	if err := dev.ReadBytes(dev.address, REG_VBAT, buf[:]); err != nil {
		return 0, 0, err
	}
	vbat = uint16(buf[0]) + (uint16(buf[1]) << 8)
	vref = uint16(buf[2]) + (uint16(buf[3]) << 8)
	vcal = uint16(buf[4]) + (uint16(buf[5]) << 8)
	Ref := 3.3 * float32(vcal) / float32(vref)
	return 2 * Ref * float32(vbat) / 4095.0, Ref, nil
}

func (dev *Device) Configuration() (ConfigurationByte, error) {
	conf, err := dev.ReadByte(dev.address, REG_CONF)
	if err != nil {
		return 0, err
	}
	return ConfigurationByte(conf), nil
}

func (dev *Device) SetConfiguration(conf ConfigurationByte) error {
	return dev.WriteByte(dev.address, REG_CONF, byte(conf))
}

func (dev *Device) ModifyConfiguration(mask ConfigurationByte, conf ConfigurationByte) error {
	return dev.ModifyByte(dev.address, REG_CONF, byte(mask), byte(conf))
}

func (dev *Device) Watchdog() (uint16, error) {
	var buf [2]byte

	err := dev.ReadBytes(dev.address, REG_WATCH, buf[:])
	if err != nil {
		return 0, err
	}
	return uint16(buf[0]) + (uint16(buf[1]) << 8), nil
}

func (dev *Device) SetWatchdog(delay uint16, conf byte) error {
	var buf [2]byte

	buf[0] = byte(delay)
	buf[1] = byte(delay >> 8)
	if err := dev.WriteBytes(dev.address, REG_WATCH, buf[:]); err != nil {
		return err
	}
	return dev.ModifyByte(dev.address, REG_CONF, conf, conf)
}

func (dev *Device) Wakeup() (uint16, error) {
	var buf [2]byte

	err := dev.ReadBytes(dev.address, REG_WAKE, buf[:])
	if err != nil {
		return 0, err
	}
	return uint16(buf[0]) + (uint16(buf[1]) << 8), nil
}

func (dev *Device) SetWakeup(delay uint16, conf byte) error {
	var buf [2]byte

	buf[0] = byte(delay)
	buf[1] = byte(delay >> 8)
	if err := dev.WriteBytes(dev.address, REG_WAKE, buf[:]); err != nil {
		return err
	}
	return dev.ModifyByte(dev.address, REG_CONF, conf, conf)
}

func (dev *Device) Alarm() (Alarm, error) {
	var buf [4]byte

	err := dev.ReadBytes(dev.address, REG_ALARM, buf[:])
	if err != nil {
		return 0, err
	}
	return Alarm(buf[0]) + (Alarm(buf[1]) << 8) + (Alarm(buf[2]) << 16) + (Alarm(buf[3]) << 24), nil
}

func (dev *Device) SetAlarm(a Alarm, conf byte) error {
	var buf [4]byte

	buf[0] = byte(a)
	buf[1] = byte(a >> 8)
	buf[2] = byte(a >> 16)
	buf[3] = byte(a >> 24)

	if err := dev.WriteBytes(dev.address, REG_ALARM, buf[:]); err != nil {
		return err
	}
	if err := dev.ModifyByte(dev.address, REG_CONF, conf, conf); err != nil {
		return err
	}
	return dev.Program(PROG_ALARM)
}

func (dev *Device) LowBatteryTimer() (uint16, error) {
    var buf [2]byte

    err := dev.ReadBytes(dev.address, REG_LBO_TIMER, buf[:])
    if err != nil {
        return 0, err
    }
    return uint16(buf[0]) + (uint16(buf[1]) << 8), nil
}

func (dev *Device) SetLowBatteryTimer(delay uint16, conf byte) error {
    var buf [2]byte

    buf[0] = byte(delay)
    buf[1] = byte(delay >> 8)
    if err := dev.WriteBytes(dev.address, REG_LBO_TIMER, buf[:]); err != nil {
        return err
    }
    return dev.ModifyByte(dev.address, REG_CONF, conf, conf)
}

