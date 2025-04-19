package config

import (
	"bytes"
	"embed"
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
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

//go:embed cities.csv
var citiesCsv string

//go:embed data
var DataFolder embed.FS

func structToTomlFile(cfg Config, filepath string) bool {
	fmt.Println(filepath)
	fmt.Println(cfg)
	BufWriter := new(bytes.Buffer)
	err := toml.NewEncoder(BufWriter).Encode(cfg)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath, BufWriter.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
	return true
}

func extractEmbeddedFolder(targetDir string) error {
	return fs.WalkDir(DataFolder, "data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel("data", path)
		destPath := filepath.Join(targetDir, "data", relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// It's a file: copy it
		data, err := DataFolder.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, 0644)
	})
}

func getConfigDir() string {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	var cfgdirpath string
	if xdg != "" {
		err := os.Mkdir(path.Join(xdg, "adhani"), 0777)
		if err != nil {
			panic(err)
		} // do nothing dir exists

		cfgdirpath = filepath.Join(xdg, "adhani")

	} else {

		home, err := os.UserHomeDir()
		if err != nil {
			panic("cannot find home directory: " + err.Error())
		}
		errr := os.Mkdir(path.Join(home, ".config", "adhani"), 0777)
		if errr != nil && !os.IsExist(errr) {
		} // do nothing dir exists

		cfgdirpath = filepath.Join(home, ".config", "adhani")
	}
	return cfgdirpath
}
func GetAppConfigDir() string {
	cfgdirpath := getConfigDir()
	if _, err := os.Stat(path.Join(cfgdirpath, "config.toml")); errors.Is(err, os.ErrNotExist) {
		structToTomlFile(gen_Default(Config{}), path.Join(cfgdirpath, "config.toml"))
	}
	extractEmbeddedFolder(cfgdirpath)
	return cfgdirpath
}

func cityLookup(city string) (geoData, error) {

	reader := csv.NewReader(strings.NewReader(citiesCsv))
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
		Latitude  float64 `toml:"latitude"`
		Longitude float64 `toml:"longitude"`
		Timezone  string  `toml:"timezone"`
	} `toml:"location"`

	Calculation struct {
		Method string `toml:"method"`
		Madhab string `toml:"madhab"`
	} `toml:"calculation"`

	Application struct {
		Timeout                    uint64 `toml:"timeout"`
		IconPath                   string `toml:"iconpath"`
		Athan                      string `toml:"athan"`
		TimeTillNextPrayerReminder bool   `toml:"timetillnextprayerreminder"`
		BeforePrayerTime           int16  `toml:"beforeprayertime"`
	} `toml:"application"`
}

var conf Config

func (c Config) IsEmpty() bool {
	return c.Location.Latitude == 0 &&
		c.Location.Longitude == 0 &&
		c.Location.Timezone == "" &&
		c.Calculation.Method == "" &&
		c.Calculation.Madhab == ""
}

func gen_Default(conf Config) Config {
	tzname, err := tzlocal.RuntimeTZ()
	if err != nil {
		log.Fatalf("error getting locale timezone : %v", err)
	}
	city := strings.Split(tzname, "/")[1]
	var geo, _ = cityLookup(city)
	fmt.Println(geo)
	long := geo.Longitude
	fmt.Println(long)
	lat := geo.Latitude
	method := "UOIF"
	madhab := "SHAFI_HANBALI_MALIKI"
	conf.Application.BeforePrayerTime = 15
	conf.Application.TimeTillNextPrayerReminder = true
	conf.Application.Timeout = 60
	conf.Calculation.Madhab = madhab
	conf.Calculation.Method = method
	conf.Location.Latitude = lat
	conf.Location.Longitude = long
	conf.Location.Timezone = tzname
	dotConfigDir := getConfigDir()
	conf.Application.Athan = path.Join(dotConfigDir, "data", "athan.mp3")
	conf.Application.IconPath = path.Join(dotConfigDir, "data", "icon.png")
	return conf
}
func LoadConfig() Config {
	fpath := path.Join(GetAppConfigDir(), "config.toml")
	if _, err := toml.DecodeFile(fpath, &conf); err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	if conf.IsEmpty() {
		gen_Default(conf)
	}
	return conf
}
