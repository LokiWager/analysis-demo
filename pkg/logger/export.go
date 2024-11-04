package logger

func Warnf(fmt string, args ...interface{}) {
	compoundSystemLogger.Warnf(fmt, args...)
}

func Debugf(fmt string, args ...interface{}) {
	compoundSystemLogger.Debugf(fmt, args...)
}

func Errorf(fmt string, args ...interface{}) {
	compoundSystemLogger.Errorf(fmt, args...)
}

func Infof(fmt string, args ...interface{}) {
	compoundSystemLogger.Infof(fmt, args...)
}

func Fatalf(fmt string, args ...interface{}) {
	compoundSystemLogger.Fatalf(fmt, args...)
}

func Panicf(fmt string, args ...interface{}) {
	compoundSystemLogger.Panicf(fmt, args...)
}

func Warn(args string) {
	compoundSystemLogger.Warn(args)
}

func Debug(args string) {
	compoundSystemLogger.Debug(args)
}

func Info(args string) {
	compoundSystemLogger.Info(args)
}

func Error(args string) {
	compoundSystemLogger.Error(args)
}

func Fatal(args string) {
	compoundSystemLogger.Fatal(args)
}

func Panic(args string) {
	compoundSystemLogger.Panic(args)
}
