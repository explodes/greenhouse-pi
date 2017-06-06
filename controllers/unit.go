package controllers

const (
	UnitStatusOn    = "on"
	UnitStatusOff   = "off"
	UnitStatusError = "error"
)

type UnitStatus string

// Unit toggles the flow of water in the system
type Unit interface {
	// Name of this Unit
	Name() string

	// On turns on the Unit
	On() error

	// Off turns off the Unit
	Off() error

	// Whether or not this Unit is on
	Status() (UnitStatus, error)
}
