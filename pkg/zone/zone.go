package zone

import (
	"time"

	log "github.com/sirupsen/logrus"

	rpio "github.com/stianeikeland/go-rpio/v4"
)

// Zone contains the interface to interact with the zone's pump and its moisture sensor
type Zone struct {
	Name               string
	PumpRelayPinNumber int
	pumpRelayPin       rpio.Pin

	interval                 time.Duration
	intervaledSupplyDuration time.Duration
	maxSupplyDurationPerDay  time.Duration

	Status Status
}

// New creates and initialzes a zone and the counters
func New(config Configuration) *Zone {
	z := Zone{
		Name:                     config.Name,
		PumpRelayPinNumber:       config.PumpRelayPinNumber,
		interval:                 time.Duration(config.IntervalSecond) * time.Second,
		intervaledSupplyDuration: time.Duration(config.IntervaledSupplySecond) * time.Second,
		maxSupplyDurationPerDay:  time.Duration(config.MaxSupplySecondsPerDay) * time.Second,
		Status:                   Status{},
	}

	z.Status.SupplyEvents = make([]SupplyEvent, supplyEventsQueueSize)

	z.pumpRelayPin = rpio.Pin(z.PumpRelayPinNumber)
	// Set pin to output mode
	z.pumpRelayPin.Output()

	return &z
}

// Supply starts the pump in the zone and turn it off after duration
// The time and duration will also be logged
func (z *Zone) Supply(duration time.Duration) {
	// The 4-relay board uses LOW to activate the relay
	log.Infof("Starting pump %s, duration %v", z.Name, duration)
	z.pumpRelayPin.Low()

	time.Sleep(duration)

	z.pumpRelayPin.High()
	log.Infof("Shutting off pump %s", z.Name)

	z.Status.SupplyEvents[z.Status.SupplyEventsHead] = SupplyEvent{time.Now(), duration}
	z.Status.SupplyEventsHead = (z.Status.SupplyEventsHead + 1) % supplyEventsQueueSize
	if z.Status.SupplyEventsHead == z.Status.SupplyEventsTail {
		z.Status.SupplyEventsTail = (z.Status.SupplyEventsTail + 1) % supplyEventsQueueSize
	}
}

// TrySupply checks if the zone can be watered for duration considering
// the mount of water supplied in the 24 hours./
// Returns true if the supply was successful, false if the max amount of water supply has been or will be exceeded with this supply.
func (z *Zone) TrySupply(duration time.Duration) bool {
	for z.Status.SupplyEventsTail != z.Status.SupplyEventsHead &&
		z.Status.SupplyEvents[z.Status.SupplyEventsTail].Time.Before(time.Now().Add(- /*24*time.Hour*/ 2*time.Minute)) {
		z.Status.SupplyEventsTail = (z.Status.SupplyEventsTail + 1) % supplyEventsQueueSize
	}

	sumDuration := time.Duration(0)
	for i := z.Status.SupplyEventsTail; i != z.Status.SupplyEventsHead; i = (i + 1) % supplyEventsQueueSize {
		sumDuration += z.Status.SupplyEvents[i].Duration
	}
	sumDuration += duration
	log.Debugf("Zone %s sumDuration: %v, maxSupplyDurationPerDay: %v", z.Name, sumDuration, z.maxSupplyDurationPerDay)

	if sumDuration > z.maxSupplyDurationPerDay {
		return false
	}

	z.Supply(duration)
	return true
}

// TryIntervaledSupply triggers an interval based water supply if the zone considering
// the mount of water supplied in the 24 hours
// Returns true if the supply is successful, false if the max amount of water supply has been or will be exceeded with this supply.
func (z *Zone) TryIntervaledSupply() bool {
	if z.Status.LastIntervaledSupply.After(time.Now().Add(-z.interval)) {
		return false
	}

	if ret := z.TrySupply(z.intervaledSupplyDuration); !ret {
		log.Infof("Zone %s interval has passed but it has reached the max supply amount per day", z.Name)
		return false
	}
	log.Infof("Zone %s has been watered for %v during an intervaled supply", z.Name, z.intervaledSupplyDuration)

	z.Status.LastIntervaledSupply = time.Now()
	z.Status.LastSupply = time.Now()
	return true
}
