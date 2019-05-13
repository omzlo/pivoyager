package main

import (
	"fmt"
	"github.com/omzlo/pivoyager/device"
	"os"
	"strconv"
	"strings"
	"time"
)

type Command struct {
	Name        string
	Execute     func(*device.Device, []string) error
	Description string
}

func assert_argc(args []string, count ...int) []string {
	argc := len(args) - 1

	if len(count) == 0 {
		return args[1:]
	}

	for _, c := range count {
		if argc == c {
			return args[1:]
		}
	}

	if len(count) == 1 {
		fmt.Fprintf(os.Stderr, "Command '%s' expects %d parameter(s), but %d were provided.\n", args[0], count[0], argc)
		os.Exit(2)
	}

	s := fmt.Sprintf("%d", count[0])
	for i := 1; i < len(count)-2; i++ {
		s += fmt.Sprintf(", %d", count[i])
	}
	s += fmt.Sprintf(" or %d", count[len(count)-1])
	fmt.Fprintf(os.Stderr, "Command '%s' expects %s parameters, but %d were provided.\n", args[0], s, argc)
	os.Exit(2)
	return nil
}

const (
	DO_FLAGS   = 1
	DO_BATTERY = 2
	DO_VOLTAGE = 4
)

func cmd_status(dev *device.Device, args []string) error {
	var todo int

	args = assert_argc(args, 0, 1)
	if len(args) == 0 {
		todo = DO_FLAGS | DO_BATTERY | DO_VOLTAGE
	} else {
		switch args[0] {
		case "battery":
			todo = DO_BATTERY
		case "flags":
			todo = DO_FLAGS
		case "voltage":
			todo = DO_VOLTAGE
		default:
			return fmt.Errorf("Unknown status type '%s'", args[0])
		}
	}
	if (todo & (DO_FLAGS | DO_BATTERY)) != 0 {
		s, err := dev.Status()
		if err != nil {
			return err
		}
		if (todo & DO_FLAGS) != 0 {
			fmt.Printf("Status: %s\n", s)
		}
		if (todo & DO_BATTERY) != 0 {
			fmt.Printf("Battery: %s\n", s.BatteryStateString())
		}
	}
	if (todo & DO_VOLTAGE) != 0 {
		vbat, vref, err := dev.Voltage()
		if err != nil {
			return err
		}
		fmt.Printf("VBat: %.2fV\n", vbat)
		fmt.Printf("VRef: %.2fV\n", vref)
	}
	return nil
}

func cmd_date(dev *device.Device, args []string) error {
	var tm time.Time
	var err error

	args = assert_argc(args, 0, 1)

	if len(args) == 0 {
		tm, err = dev.Time()
		if err != nil {
			return err
		}
		fmt.Println(tm.Format(time.RFC3339))
	} else {
		if args[0] != "sync" {
			tm, err = time.Parse(time.RFC3339, args[0])
			if err != nil {
				return fmt.Errorf("Failed to parse date: %s", err)
			}
		} else {
			for {
				tm = time.Now()
				if tm.Nanosecond() < 100000 {
					break
				}
				time.Sleep(50000 * time.Nanosecond)
			}
		}
		dev.SetTime(tm.UTC())
		fmt.Printf("Setting date to %s\n", tm.UTC())
	}
	return nil
}

/*
func cmd_voltage(dev *device.Device, args []string) error {
	args = assert_argc(args, 0)

	vbat, vref, err := dev.Voltage()
	if err != nil {
		return err
	}
	fmt.Printf("VBat: %.2fV\n", vbat)
	fmt.Printf("VRef: %.2fV\n", vref)
	return nil
}
*/

func cmd_watchdog(dev *device.Device, args []string) error {
	args = assert_argc(args, 0, 1)

	if len(args) == 1 {
		delay, err := strconv.ParseUint(args[0], 0, 16)
		if err != nil {
			return err
		}
		if err := dev.SetWatchdog(uint16(delay), device.CONF_I2C_WD); err != nil {
			return err
		}
		fmt.Println("OK")
	} else {
		delay, err := dev.Watchdog()
		if err != nil {
			return err
		}
		conf, err := dev.Configuration()
		if err != nil {
			return err
		}
		conf = conf & 0x3
		fmt.Printf("Watchdog: %d seconds\n", delay)
		options := conf.ToStrings()
		var res string
		if len(options) == 0 {
			res = "disabled"
		} else {
			res = strings.Join(options, " ")
		}
		fmt.Printf("Options: %s\n", res)
	}
	return nil
}

func cmd_wakeup(dev *device.Device, args []string) error {
	args = assert_argc(args, 0, 1)

	if len(args) == 1 {
		delay, err := strconv.ParseUint(args[0], 0, 16)
		if err != nil {
			return err
		}

		if err := dev.SetWakeup(uint16(delay), device.CONF_WAKE_AFTER); err != nil {
			return err
		}
		fmt.Println("OK")
	} else {
		delay, err := dev.Wakeup()
		if err != nil {
			return err
		}
		conf, err := dev.Configuration()
		if err != nil {
			return err
		}
		conf = conf & 0x3C
		fmt.Printf("Wakeup: %d seconds\n", delay)
		options := conf.ToStrings()
		var res string
		if len(options) == 0 {
			res = "disabled"
		} else {
			res = strings.Join(options, " ")
		}
		fmt.Printf("Options: %s\n", res)
	}
	return nil
}

/*
func cmd_battery(dev *device.Device, args []string) error {
	s, err := dev.Status()
	if err != nil {
		return err
	}
	fmt.Println(s.BatteryStateString())
	return nil
}
*/

func cmd_alarm(dev *device.Device, args []string) error {
	args = assert_argc(args, 0, 1)

	if len(args) == 1 {
		var alarm device.Alarm

		if err := alarm.UnmarshalText([]byte(args[0])); err != nil {
			return err
		}
		if err := dev.SetAlarm(alarm, device.CONF_WAKE_ALARM); err != nil {
			return err
		}
		fmt.Println("OK")
	} else {
		alarm, err := dev.Alarm()
		if err != nil {
			return err
		}
		fmt.Printf("Alarm: %s (%x)\n", alarm, uint32(alarm))
	}
	return nil
}

func cmd_enable(dev *device.Device, args []string) error {
	args = args[1:]
	if len(args) == 0 {
		conf, err := dev.Configuration()
		if err != nil {
			return err
		}
		fmt.Printf("Enabled: %s\n", conf)
	} else {
		var options device.ConfigurationByte
		if err := options.FromStrings(args); err != nil {
			return err
		}
		if err := dev.ModifyConfiguration(options, options); err != nil {
			return err
		}
		fmt.Println("OK")
	}
	return nil
}

func cmd_disable(dev *device.Device, args []string) error {
	args = args[1:]
	if len(args) == 0 {
		conf, err := dev.Configuration()
		if err != nil {
			return err
		}
		fmt.Printf("Disabled: %s\n", conf.Invert())
	} else {
		var options device.ConfigurationByte
		if err := options.FromStrings(args); err != nil {
			return err
		}
		if err := dev.ModifyConfiguration(options, 0); err != nil {
			return err
		}
		fmt.Println("OK")
	}
	return nil
}

func cmd_clear(dev *device.Device, args []string) error {
	args = assert_argc(args, 1, 1)

	switch args[0] {
	case "alarm":
		if err := dev.Program(device.PROG_CLEAR_ALARM); err != nil {
			return err
		}
	case "button":
		if err := dev.Program(device.PROG_CLEAR_BUTTON); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown flag '%s' for clear command.", args[0])
	}
	fmt.Println("OK")
	return nil
}

var commands = []Command{
	Command{"alarm", cmd_alarm, `Get current alarm date, or set it (alarm <alarm-pattern>).
                The format of <alarm-pattern> is day-hour-minute-second, where:
                - day is "Mon", "Tue", ... "Sun" to select a day of the week.
                - day is 1-31 to identify a day of the month
                - day is "*" to ignore the day
                - hour is 0-24 to select an hour, or "*" to ignore
                - minute is 0-59 to select a minute or "*" to ignore
                - second is 0-59 to select a second or "*" to ignore
	`},
	Command{"clear", cmd_clear, `Clear a status bit (clear <flags>).
				The value <flags> can be either "button" or "alarm".
	`},
	Command{"date", cmd_date, `Get the current RTC time, or set it (date <utc-time-RFC3339>.
                Use 'date sync' to use the current operating system date for the RTC.
                Time is typically expressed as UTC time to avoid any ambiguity.
                See RFC3339 for a valid date format.
	`},
	Command{"disable", cmd_disable, `Disable configuration options for watchdog and wakeup.
                See 'enable' command for options.
	`},
	Command{"enable", cmd_enable, `Enable configuration options for watchdog and wakeup
                Options are:
                - "i2c-watchdog" enable i2c based watchdog.
                - "gpio-watchdog" enable GPIO 26 watchdog (pin 37 on 40-pin header).
                - "timer-wakeup" enable wakeup after timer, as set with wakeup command.
                - "alarm-wakeup" enable wakeup on alarm.
                - "power-wakeup" wakeup if USB power goes up (only if it was down during shutdown).
                - "button-wakeup" wakeup if user presses button.
				Note: "timer-wakeup" cancels "alarm-wakeup".
	`},
	Command{"help", nil, `Prints this message.
	`},
	Command{"status", cmd_status, `Get the current UPS status of the PiVoyager.
				- "status flags" shows system status flags.
				- "status battery" shows battery status (e.g. "charging").
				- "status volatge" shows battery and reference voltage.
				- "status" shows all of the above.
	`},
	Command{"wakeup", cmd_wakeup, `Get wakeup information, or set wakeup time (wakeup <seconds>)
				Note: "wakeup" sets an alarm, overriding any alarm previously set.
	`},
	Command{"watchdog", cmd_watchdog, `Get watchdog information, or set watchdog time (watchdog <seconds>)
	`},
}

var PIVOYAGER_VERSION = "0.1"

func version() {
	fmt.Printf("This is pivoyager version %s, (c) Omzlo P.C. [omzlo.com]\n", PIVOYAGER_VERSION)
}

func help() {
	version()
	fmt.Println("Syntax: pivoyager <command> (options...)")
	fmt.Println("Valid commands are:")
	for _, command := range commands {
		fmt.Printf("  %10s:  %s\n", command.Name, command.Description)
	}
}

func main() {

	if len(os.Args) == 1 {
		version()
		fmt.Println("Type 'pivoyager help' for usage information.")
		os.Exit(0)
	}

	if os.Args[1] == "help" {
		help()
		os.Exit(0)
	}

	for _, command := range commands {
		if command.Name == os.Args[1] {
			pivoyager, err := device.Open(false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to connect to pivoyager.\n")
				fmt.Fprintf(os.Stderr, "Could not connect to i2c device: %s\n", err)
				os.Exit(1)
			}
			if err := command.Execute(pivoyager, os.Args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	fmt.Fprintf(os.Stderr, "Error: command '%s' unknown\n", os.Args[1])
}
