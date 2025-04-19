package main

import (
	"adhani/config"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"

	calc "github.com/mnadev/adhango/pkg/calc"
	data "github.com/mnadev/adhango/pkg/data"
	util "github.com/mnadev/adhango/pkg/util"
)

func beforeprayerreminder(conf config.Config, prayers []Prayer) {
	for {
		for _, prayer := range prayers {
			now := int64(time.Now().Unix())
			time_left := int16((prayer.Time - now) / 60)
			if time_left > 0 && time_left == conf.Application.BeforePrayerTime {
				var time_string string = strconv.Itoa(int(conf.Application.BeforePrayerTime))
				var m Message = Message{Urgency: "normal", Message: "Prayer Time in " + time_string + " minutes\n" + prayer.Prayer + " time is in " + time_string + " minutes!", Icon: conf.Application.IconPath}
				go notify(m)
			}

		}
		time.Sleep(time.Duration(conf.Application.Timeout) * time.Second)
	}

}

type Message struct {
	Urgency string
	Message string
	Icon    string
	Athan   *string
}

func notify(message Message) {
	cmd := exec.Command(
		"notify-send",
		"--urgency="+message.Urgency,
		"--app-name='adhani'",
		"--icon="+message.Icon,
		message.Message,
		"--hint=int:transient:1")
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	if message.Athan != nil { // no athan
		cmd = exec.Command("paplay", *message.Athan)
		err = cmd.Run()
		if err != nil {
			log.Fatalf("Failed to play sound: %v", err)
		}
	}
}

func calculatePrayers(conf config.Config) *calc.PrayerTimes {
	date := data.NewDateComponents(time.Now())
	var params *calc.CalculationParameters
	switch {
	case conf.Calculation.Method == "MUSLIM_WORLD_LEAGUE":
		params = calc.GetMethodParameters(calc.MUSLIM_WORLD_LEAGUE)
	case conf.Calculation.Method == "EGYPTIAN":
		params = calc.GetMethodParameters(calc.EGYPTIAN)
	case conf.Calculation.Method == "KARACHI":
		params = calc.GetMethodParameters(calc.KARACHI)
	case conf.Calculation.Method == "UMM_AL_QURA":
		params = calc.GetMethodParameters(calc.UMM_AL_QURA)
	case conf.Calculation.Method == "DUBAI":
		params = calc.GetMethodParameters(calc.DUBAI)
	case conf.Calculation.Method == "MOON_SIGHTING_COMMITTEE":
		params = calc.GetMethodParameters(calc.MOON_SIGHTING_COMMITTEE)
	case conf.Calculation.Method == "NORTH_AMERICA":
		params = calc.GetMethodParameters(calc.NORTH_AMERICA)
	case conf.Calculation.Method == "KUWAIT":
		params = calc.GetMethodParameters(calc.KUWAIT)
	case conf.Calculation.Method == "QATAR":
		params = calc.GetMethodParameters(calc.QATAR)
	case conf.Calculation.Method == "SINGAPORE":
		params = calc.GetMethodParameters(calc.SINGAPORE)
	case conf.Calculation.Method == "UOIF":
		params = calc.GetMethodParameters(calc.UOIF)
	}

	coords, err := util.NewCoordinates(conf.Location.Latitude, conf.Location.Longitude)
	if err != nil {
		fmt.Printf("got error %+v", err)
		panic(err)

	}

	prayerTimes, err := calc.NewPrayerTimes(coords, date, params)
	if err != nil {
		fmt.Printf("got error %+v", err)
		panic(err)
	}

	err = prayerTimes.SetTimeZone(conf.Location.Timezone)
	if err != nil {
		fmt.Printf("got error %+v", err)
		panic(err)

	}

	return prayerTimes
}

type Prayer struct {
	Time   int64
	Prayer string
}

func main() {
	var conf config.Config = config.LoadConfig()

	var timeray []Prayer
	updateTimeray := func() {
		times := calculatePrayers(conf)
		timeray = []Prayer{
			{Time: times.Fajr.Unix(), Prayer: "fajr"},
			{Time: times.Dhuhr.Unix(), Prayer: "dhuhr"},
			{Time: times.Asr.Unix(), Prayer: "asr"},
			{Time: times.Maghrib.Unix(), Prayer: "maghrib"},
			{Time: times.Isha.Unix(), Prayer: "isha"},
		}
	}

	// once at startup
	updateTimeray()
	lastUpdate := time.Now()

	if conf.Application.TimeTillNextPrayerReminder {
		go beforeprayerreminder(conf, timeray)
	}

	for {
		// every 30 minutes, recalc
		if time.Since(lastUpdate) >= 30*time.Minute {
			updateTimeray()
			lastUpdate = time.Now()
		}

		now := time.Now().Unix()
		for _, p := range timeray {
			if (now > p.Time && now-int64(conf.Application.Timeout) < p.Time) || now == p.Time {
				m := Message{
					Urgency: "normal",
					Message: fmt.Sprintf("Time for %s prayer.", p.Prayer),
					Icon:    conf.Application.IconPath,
					Athan:   &conf.Application.Athan,
				}
				go notify(m)
			}
		}

		// sleep between checks
		time.Sleep(time.Duration(conf.Application.Timeout) * time.Second)
	}
}
