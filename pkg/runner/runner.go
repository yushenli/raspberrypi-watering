package runner

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/yushenli/raspberrypi-watering/pkg/zone"
)

// Runner periodically checks all the zones and supply water at their configured interval or when they are dry
type Runner struct {
	// Zones is a map containing all the zones to be checked
	Zones map[string]*zone.Zone

	CheckInterval      time.Duration
	StatusJSONFilename string
}

// Run starts the main execution loop
func (r *Runner) Run() {
	for true {
		supplied := false
		for _, zone := range r.Zones {
			time.Sleep(1 * time.Second)
			log.Infof("Checking zone %s", zone.Name)
			supplied = supplied || zone.TryIntervaledSupply()
		}

		// If at least one zone is watered in this cycle, persist the statuses
		if supplied {
			statuses := make(map[string]zone.StatusJSON)
			for name, zone := range r.Zones {
				statuses[name] = zone.Status.Compact()
			}

			if err := zone.PersistStatusesToJSON(statuses, r.StatusJSONFilename); err != nil {
				log.Errorf("Failed to persist statuses to JSON file %s: %v", r.StatusJSONFilename, err)
			} else {
				log.Infof("Persisted statuses of %d zones to JSON file %s", len(r.Zones), r.StatusJSONFilename)
			}
		}

		time.Sleep(r.CheckInterval)
	}
}
