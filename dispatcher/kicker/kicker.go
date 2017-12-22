package kicker

// Kicker is an interface to control frequency of kicking a Kickable
// object.
type Kicker interface {
	Start(kickable Kickable)
	Stop() <-chan struct{}
	Ping()
	PollingInterval() uint
}

// Config is a builder of a Kicker.
type Config interface {
	NewKicker() Kicker
}

// Kickable is an interface of something kickable.
type Kickable interface {
	Kick()
}
