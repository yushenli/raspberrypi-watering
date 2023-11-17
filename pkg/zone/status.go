package zone

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

const (
	// supplyEventsQueueSize is the size of the circular queue that holds the last supply events
	supplyEventsQueueSize = 100
)

// SupplyEvent represents an occurence of a water supply. It contains the time and duration of the supply.
type SupplyEvent struct {
	Time     time.Time
	Duration time.Duration
}

// Status holds the status of a zone, including the recent supplying history
type Status struct {
	LastSupply           time.Time
	LastIntervaledSupply time.Time

	// SupplyEvents is a circular queue which contains the last supplyEventsQueueSize supply events
	SupplyEvents     []SupplyEvent
	SupplyEventsHead int
	SupplyEventsTail int
}

// StatusJSON contains the core data of an Status object for serialization/deserialization
type StatusJSON struct {
	LastSupply           time.Time `json:"last_supply"`
	LastIntervaledSupply time.Time `json:"last_intervaled_supply"`
	SupplyEvents         []SupplyEvent
}

// Compact converts a Status object to a StatusJSON object
func (s Status) Compact() StatusJSON {
	ret := StatusJSON{
		LastSupply:           s.LastSupply,
		LastIntervaledSupply: s.LastIntervaledSupply,
		SupplyEvents:         nil,
	}

	for i := s.SupplyEventsTail; i != s.SupplyEventsHead; i = (i + 1) % supplyEventsQueueSize {
		ret.SupplyEvents = append(ret.SupplyEvents, s.SupplyEvents[i])
	}

	return ret
}

// Unpack adopts the data from a StatusJSON object to a Status object
func (s *Status) Unpack(sj StatusJSON) {
	s.LastSupply = sj.LastSupply
	s.LastIntervaledSupply = sj.LastIntervaledSupply
	for _, e := range sj.SupplyEvents {
		s.SupplyEvents[s.SupplyEventsHead] = e
		s.SupplyEventsHead = (s.SupplyEventsHead + 1) % supplyEventsQueueSize
		if s.SupplyEventsHead == s.SupplyEventsTail {
			s.SupplyEventsTail = (s.SupplyEventsTail + 1) % supplyEventsQueueSize
		}
	}
}

// PersistStatusesToJSON serializes a map of StatusJSON objects to a JSON file.
func PersistStatusesToJSON(statuses map[string]StatusJSON, filename string) error {
	data, err := json.Marshal(statuses)
	if err != nil {
		return fmt.Errorf("Failed to marshal statuses to JSON: %v", err)
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// DeserializeStatusesFromJSON deserializes a JSON file into a map of StatusJSON objects.
func DeserializeStatusesFromJSON(filename string) (map[string]StatusJSON, error) {
	var statuses map[string]StatusJSON
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from JSON file %s: %v", filename, err)
	}

	err = json.Unmarshal(data, &statuses)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshall statuses from JSON file %s: %v", filename, err)
	}

	return statuses, nil
}
