package main

import (
	"container/ring"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"
)

// Config file structure type
type Config struct {
	Sensors      []string `yaml:"sensors"`
	Fan          string   `yaml:"fan"`
	Mapping      []string `yaml:"mapping"`
	DefaultState string   `yaml:"defaultstate"`
	BufferSize   int      `yaml:"buffersize"`
	Matrix       []struct {
		Temp  int `yaml:"temp"`
		Level int `yaml:"level"`
	} `yaml:"matrix"`
}

// FanObj - fan data struct
type FanObj struct {
	Level  string `yaml:"level"`
	Speed  string `yaml:"speed"`
	Status string `yaml:"status"`
}

// MaxIntSlice function returns Max int from the specified Int slice
func MaxIntSlice(v []int) int {
	sort.Ints(v)
	return v[len(v)-1]
}

// checkProcessError function checks err variable, terminates execution and returns corresponding exitcode
func checkProcessError(err error, exitcode int) {
	if err != nil {
		fmt.Println(err)
		os.Exit(exitcode)
	}
}

//
func setFanLevel(cfg *Config, fan *FanObj, tempReading int) {
	var level int
	for _, currentMatrix := range cfg.Matrix {
		if currentMatrix.Temp != -1 {
			if tempReading >= currentMatrix.Temp {
				level = currentMatrix.Level
			}
		}
	}
	levelReading, _ := strconv.Atoi(fan.Level)
	if level != levelReading {
		writeFan(cfg, level, 8)
	}
	err := ioutil.WriteFile(cfg.Fan, []byte("watchdog 3"), 0644)
	checkProcessError(err, 13)
}

// readFile function exports configuration taken from config.yml to a Config structure
func readFile(cfg *Config) {
	f, err := os.Open("/etc/thinkfancontrol/config.yml")
	checkProcessError(err, 2)
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	checkProcessError(err, 3)
}

func readFanStatus(FanObj *FanObj, fanPath string) {
	f, err := os.Open(fanPath)
	checkProcessError(err, 5)
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(FanObj)
	checkProcessError(err, 6)
}

func writeFan(cfg *Config, level int, errorCode int) {
	levelString := strconv.Itoa(level)
	d1 := []byte("")
	if level == -1 {
		log.Println("Setting fan level: auto")
		d1 = []byte("level auto")
	} else {
		log.Println("Setting fan level: " + levelString)
		d1 = []byte("level " + levelString)
	}
	err := ioutil.WriteFile(cfg.Fan, d1, 0644)
	checkProcessError(err, errorCode)
}

func main() {

	var cfg Config
	var fan FanObj
	var sensorsList []string

	readFile(&cfg)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		for sig := range c {
			log.Println(sig)
			writeFan(&cfg, -1, 8)
			log.Println("exiting...")
			os.Exit(0)
		}
	}()

	writeFan(&cfg, 0, 9)

	if cfg.BufferSize == 0 {
		cfg.BufferSize = 10
	}
	ringBuffer := ring.New(cfg.BufferSize)

	for _, singleSensorPath := range cfg.Sensors {
		sensors, err := filepath.Glob(singleSensorPath)
		checkProcessError(err, 4)
		sensorsList = append(sensorsList, sensors...)
	}
	for {
		var sensorReadingsInt = []int{}
		var sensorReadingIntAvg int

		for _, sensor := range sensorsList {
			sensorReading, _ := ioutil.ReadFile(sensor)
			var sensorReadingString = strings.TrimSuffix(string(sensorReading), "\n")
			if sensorReadingString != "" {
				i, err := strconv.Atoi(sensorReadingString)
				checkProcessError(err, 5)
				i = i / 1000
				sensorReadingsInt = append(sensorReadingsInt, i)

			}
		}
		maxSensorReading := MaxIntSlice(sensorReadingsInt)
		ringBuffer.Value = maxSensorReading
		ringBuffer = ringBuffer.Next()
		var bufferSize int
		ringBuffer.Do(func(x interface{}) {
			if x != nil {
				bufferSize++
				sensorReadingIntAvg = sensorReadingIntAvg + x.(int)
			}
		})
		tempReading := sensorReadingIntAvg / bufferSize
		readFanStatus(&fan, cfg.Fan)

		setFanLevel(&cfg, &fan, tempReading)

		time.Sleep(1 * time.Second)
	}
	select {}
}
