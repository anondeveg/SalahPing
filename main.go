package main

import (
	"adhani/config"
	"fmt"
	"os/exec"
	"time"

	calc "github.com/mnadev/adhango/pkg/calc"
	data "github.com/mnadev/adhango/pkg/data"
	util "github.com/mnadev/adhango/pkg/util"
)

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
	default:
		// fallback to a sensible default
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

func main() {
	var conf config.Config = config.LoadConfig()

	times := calculatePrayers(conf)
	for {
		time.Sleep(60 * time.Second)
		isnow := int64(times.NextPrayerNow()) == 0
		if isnow {
			cmd := exec.Command(
				"notify-send",
				"--urgency=critical", // make it stand outreturn
				"It's time for Fajr. Wake up and pray 🙏",
			)
			err := cmd.Run()
			if err != nil {
				panic(err)
			}

		}

	}

}
