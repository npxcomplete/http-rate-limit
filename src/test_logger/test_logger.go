package test_logger

type LineLogger struct {
	Lines []string
}

func (log *LineLogger) Error(msg string) {
	log.Lines = append(log.Lines, msg)
}

type NoopLogger struct {
}

func (log NoopLogger) Error(msg string) {
}
