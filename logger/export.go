package logs

// Sync calls *zap.Logger.Sync
func Sync() error {
	var err error

	if l != nil {
		err = l.sync()
	}
	return err
}

// Debugf calls *zap.SugaredLogger.Debugf
func Debugf(template string, args ...interface{}) {
	if l != nil {
		l.debugf(template, args...)
	}
}

// Infof calls *zap.SugaredLogger.Infof
func Infof(template string, args ...interface{}) {
	if l != nil {
		l.infof(template, args...)
	}
}

// Warnf calls *zap.SugaredLogger.Warnf
func Warnf(template string, args ...interface{}) {
	if l != nil {
		l.warnf(template, args...)
	}
}

// Errorf calls *zap.SugaredLogger.Errorf
func Errorf(template string, args ...interface{}) {
	if l != nil {
		l.errorf(template, args...)
	}
}

// DPanicf calls *zap.SugaredLogger.DPanicf
func DPanicf(template string, args ...interface{}) {
	if l != nil {
		l.dpanicf(template, args...)
	}
}
