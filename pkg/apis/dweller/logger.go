package dweller

// Logger interface is a contract for logger being used by dweller.
type Logger interface {
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}

// dummyLogger is a dummy implementation of Logger contract.
type dummyLogger struct{}

// Infof does nothing, it's a dummy object
func (*dummyLogger) Infof(string, ...interface{}) {}

// Warnf does nothing, it's a dummy object
func (*dummyLogger) Warnf(string, ...interface{}) {}

// Errorf does nothing, it's a dummy object
func (*dummyLogger) Errorf(string, ...interface{}) {}
