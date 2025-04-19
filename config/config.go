package config

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/thlib/go-timezone-local/tzlocal"
)

type geoData struct {
	Country   string
	City      string
	Longitude float64
	Latitude  float64
}

func cityLookup(city string) (geoData, error) {
	file, err := os.Open("cities.csv")
	if err != nil {
		return geoData{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return geoData{}, err
	}
	for i, row := range records {
		if i == 0 {
			continue
		}
		if strings.EqualFold(row[0], city) { // EqualFold compares two strings in a case-insensitive manner
			long, _ := strconv.ParseFloat(row[3], 64)
			lat, _ := strconv.ParseFloat(row[2], 64)
			return geoData{City: row[0], Country: row[1], Longitude: long, Latitude: lat}, nil
		}
	}
	return geoData{}, nil
}

type Config struct {
	Location struct {
		Latitude  float64
		Longitude float64
		Timezone  string
	}
	Calculation struct {
		Method string
		Madhab string
	}
	Application struct {
		Timeout                    uint64
		IconPath                   string
		Athan                      string
		TimeTillNextPrayerReminder bool
		BeforePrayerTime           int16
	}
}

var conf Config

func (c Config) IsEmpty() bool {
	return c.Location.Latitude == 0 &&
		c.Location.Longitude == 0 &&
		c.Location.Timezone == "" &&
		c.Calculation.Method == "" &&
		c.Calculation.Madhab == ""
}

func gen_Default() {
	tzname, err := tzlocal.RuntimeTZ()
	if err != nil {
		log.Fatalf("error getting locale timezone : %v", err)
	}
	city := strings.Split(tzname, "/")[1]
	var geo, _ = cityLookup(city)
	long := geo.Longitude
	lat := geo.Latitude
	method := "UOIF"
	madhab := "SHAFI_HANBALI_MALIKI"
	conf.Calculation.Madhab = madhab
	conf.Calculation.Method = method
	conf.Location.Latitude = lat
	conf.Location.Longitude = long
	conf.Location.Timezone = tzname
}
func LoadConfig() Config {

	if _, err := toml.DecodeFile("config.toml", &conf); err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	if conf.IsEmpty() {
		gen_Default()
	}
	return conf
}
