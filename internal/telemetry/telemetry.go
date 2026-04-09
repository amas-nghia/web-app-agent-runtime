package telemetry

type Logger interface {
	Log(msg string, args ...any)
}
