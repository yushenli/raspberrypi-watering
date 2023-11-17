package zone

// Configuration holds the configs for a zone, including the water supplying interval and amount setting
type Configuration struct {
	// Name represents the zone name
	Name string `json:"name"`

	// IntervalSecond represents the scheduler watering interval.
	// The interval is represented in interger of seconds to make JSON serialization/deserialization easier
	// Set to non-positive number to disable intervaled watering
	IntervalSecond int `json:"interval_second"`

	// IntervaledSupplySecond represents how long, in second, should water be supplied each time
	// an interval-based supply takes place.
	IntervaledSupplySecond int `json:"interval_supply_second"`

	// MaxSupplySecondsPerDay represents the maximum amount of time, in second,
	// that the zone can be supplied with water in a day.
	MaxSupplySecondsPerDay int `json:"max_supply_second_per_day"`

	// PumpRelayPinNumber represents the GPIO pin number that the pump relay is connected to
	PumpRelayPinNumber int `json:"pump_relay_pin_number"`
}
