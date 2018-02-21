package log

// Logger interface is a contract for logger being used by dweller.
type Logger interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}

// Dummy is a dummy implementation of Logger contract.
type Dummy struct{}

// Debugf does nothing, it's a dummy object
func (*Dummy) Debugf(string, ...interface{}) {}

// Infof does nothing, it's a dummy object
func (*Dummy) Infof(string, ...interface{}) {}

// Warnf does nothing, it's a dummy object
func (*Dummy) Warnf(string, ...interface{}) {}

// Errorf does nothing, it's a dummy object
func (*Dummy) Errorf(string, ...interface{}) {}
