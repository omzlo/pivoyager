package device

import (
	"fmt"
	"strconv"
	"strings"
)

type Alarm uint32

const (
	AL_SECOND         Alarm = 0
	AL_SECOND_MASK          = 7
	AL_MINUTE               = 8
	AL_MINUTE_MASK          = 15
	AL_HOUR                 = 16
	AL_HOUR_MASK            = 23
	AL_DAY                  = 24
	AL_WEEKDAY_SELECT       = 30
	AL_DAY_MASK             = 31
	AL_DONT_CARE      Alarm = (1 << AL_DAY_MASK) | (1 << AL_HOUR_MASK) | (1 << AL_MINUTE_MASK) | (1 << AL_SECOND_MASK)
)

var weekdays = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

func BCDstring(b Alarm) string {
	return string('0' + FromBCD(byte(b&0xF)))
}

func MewAlarm() *Alarm {
	a := new(Alarm)
	*a = AL_DONT_CARE
	return a
}

func (a *Alarm) OnDay(day int) *Alarm {
	*a &= ^Alarm(AL_WEEKDAY_SELECT | (0x3F << AL_DAY) | (1 << AL_DAY_MASK))
	*a |= (Alarm(ToBCD(day)) << AL_DAY)
	return a
}

func (a *Alarm) OnWeekday(day int) *Alarm {
	*a &= ^Alarm((0x3F << AL_DAY) | (1 << AL_DAY_MASK))
	*a |= (Alarm(ToBCD(day)) << AL_DAY) | (1 << AL_WEEKDAY_SELECT)
	return a
}

func (a *Alarm) OnHour(hour int) *Alarm {
	*a &= ^Alarm((0x3F << AL_HOUR) | (1 << AL_HOUR_MASK))
	*a |= (Alarm(ToBCD(hour)) << AL_HOUR)
	return a
}

func (a *Alarm) OnMinute(minute int) *Alarm {
	*a &= ^Alarm((0x3F << AL_MINUTE) | (1 << AL_MINUTE_MASK))
	*a |= (Alarm(ToBCD(minute)) << AL_MINUTE)
	return a
}

func (a *Alarm) OnSecond(second int) *Alarm {
	*a &= ^Alarm((0x3F << AL_SECOND) | (1 << AL_SECOND_MASK))
	*a |= (Alarm(ToBCD(second)) << AL_SECOND)
	return a
}

func (a Alarm) String() string {
	var s string

	if (a & (1 << AL_DAY_MASK)) != 0 {
		s = "*-"
	} else {
		if (a & (1 << AL_WEEKDAY_SELECT)) == 0 {
			day := (a >> AL_DAY) & 0x3F
			s = BCDstring(day>>4) + BCDstring(day) + "-"
		} else {
			day := (a >> AL_DAY) & 0xF
			s = weekdays[day-1][:3] + "-"
		}
	}

	if (a & (1 << AL_HOUR_MASK)) != 0 {
		s += "*-"
	} else {
		hour := (a >> AL_HOUR) & 0x3F
		s += BCDstring(hour>>4) + BCDstring(hour) + "-"
	}

	if (a & (1 << AL_MINUTE_MASK)) != 0 {
		s += "*-"
	} else {
		minute := (a >> AL_MINUTE) & 0x7F
		s += BCDstring(minute>>4) + BCDstring(minute) + "-"
	}

	if (a & (1 << AL_SECOND_MASK)) != 0 {
		s += "*"
	} else {
		second := (a >> AL_SECOND) & 0x7F
		s += BCDstring(second>>4) + BCDstring(second)
	}
	return s
}

func matchWeekday(s string) (int, error) {
	if len(s) < 3 {
		return 0, fmt.Errorf("Weekday name must be at least 3 characters long, got %d.", len(s))
	}
	candidate := strings.Title(s)
	for k, v := range weekdays {
		if strings.HasPrefix(candidate, v) {
			return k + 1, nil
		}
	}
	return 0, fmt.Errorf("'%s' does not match a weekday name.", s)
}

func (a *Alarm) UnmarshalText(data []byte) error {
	part := strings.Split(string(data), "-")
	if len(part) != 4 {
		return fmt.Errorf("Alarm string should be in 4 parts seperated by '-' (found %d parts)", len(part))
	}
	*a = AL_DONT_CARE
	if part[0] != "*" {
		if part[0][0] >= '0' && part[0][0] <= '9' {
			day, err := strconv.Atoi(part[0])
			if err != nil {
				return err
			}
			if day > 31 || day < 1 {
				return fmt.Errorf("Invalid day value: %d", day)
			}
			a.OnDay(day)
		} else {
			day, err := matchWeekday(part[0])
			if err != nil {
				return err
			}
			a.OnWeekday(day)
		}
	}

	if part[1] != "*" {
		hour, err := strconv.Atoi(part[1])
		if err != nil {
			return err
		}
		if hour > 24 || hour < 0 {
			return fmt.Errorf("Invalid hour value: %d", hour)
		}
		a.OnHour(hour)
	}

	if part[2] != "*" {
		minute, err := strconv.Atoi(part[2])
		if err != nil {
			return err
		}
		if minute > 59 || minute < 0 {
			return fmt.Errorf("Invalid minute value: %d", minute)
		}
		a.OnMinute(minute)
	}

	if part[3] != "*" {
		second, err := strconv.Atoi(part[3])
		if err != nil {
			return err
		}
		if second > 59 || second < 0 {
			return fmt.Errorf("Invalid second value: %d", second)
		}
		a.OnSecond(second)
	}
	return nil
}
