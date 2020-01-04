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
	Database     struct {
		Username string `yaml:"user"`
		Password string `yaml:"pass"`
	} `yaml:"database"`
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

// readFile function exports configuration taken from config.yml to a Config structure
func readFile(cfg *Config) {
	f, err := os.Open("config.yml")
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

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			log.Println(sig)
			log.Println("exiting...")
			os.Exit(0)
		}
	}()

	var cfg Config
	var fan FanObj
	var sensorsList []string

	readFile(&cfg)

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

				//log.Println(strconv.Itoa(i))
			}
		}
		maxSensorReading := MaxIntSlice(sensorReadingsInt)
		// log.Println("-----------------------------")
		// log.Println(strconv.Itoa(maxSensorReading))
		// log.Println("-----------------------------")
		ringBuffer.Value = maxSensorReading
		ringBuffer = ringBuffer.Next()

		var bufferSize int
		ringBuffer.Do(func(x interface{}) {
			if x != nil {
				bufferSize++
				//fmt.Println(x)
				sensorReadingIntAvg = sensorReadingIntAvg + x.(int)
			}
		})
		log.Println(strconv.Itoa(sensorReadingIntAvg / bufferSize))
		// fmt.Println(ringBuffer.Len())
		if bufferSize == cfg.BufferSize {
			log.Println("Buffer ready")
		} else {
			log.Println("Buffer loading")
		}
		readFanStatus(&fan, cfg.Fan)
		log.Println(fan.Level)
		log.Println(fan.Speed)
		log.Println(fan.Status)

		time.Sleep(1 * time.Second)
	}
	select {}
}