package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stianeikeland/go-rpio/v4"

	"github.com/yushenli/raspberrypi-watering/pkg/runner"
	"github.com/yushenli/raspberrypi-watering/pkg/zone"
)

func loadZonesConfig(filename string, runner *runner.Runner) error {
	var configs []zone.Configuration

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Failed to read zone configs from JSON file %s: %v", filename, err)
	}

	err = json.Unmarshal(data, &configs)
	if err != nil {
		return fmt.Errorf("Failed to unmarshall zone configs from JSON file %s: %v", filename, err)
	}

	for _, config := range configs {
		runner.Zones[config.Name] = zone.New(config)
	}
	log.Infof("Loaded %d zone configs from JSON file %s:\n%+v", len(configs), filename, configs)

	return nil
}

func main() {
	zoneConfigFilename := flag.String(
		"zone-config-filename", "rpi_zone_configs.json", "The filename for the zones config JSON file")
	statusJSONFilename := flag.String(
		"status-filename", "rpi_water_status.json", "The filename for storing the execution status JSON data")
	checkInterval := flag.Duration("check-interval", 5*time.Minute, "The interval between each round of zone check")

	// Parse the flags
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	// Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()

	r := runner.Runner{
		Zones:              make(map[string]*zone.Zone),
		CheckInterval:      *checkInterval,
		StatusJSONFilename: *statusJSONFilename,
	}
	if err := loadZonesConfig(*zoneConfigFilename, &r); err != nil {
		log.Fatalf("Failed to load zone configs from JSON file %s: %v", *zoneConfigFilename, err)
	}

	statuses, err := zone.DeserializeStatusesFromJSON(*statusJSONFilename)
	if err != nil {
		log.Warnf("JSON status file %s not existed or failed to load, starting with empty status", *statusJSONFilename)
	} else {
		for name, status := range statuses {
			if _, ok := r.Zones[name]; ok {
				r.Zones[name].Status.Unpack(status)
				log.Infof("Loaded status for zone %s", name)
			}
		}
	}

	r.Run()
}
