package aiusage

import "context"

type Record struct {
	UserID    int64
	Operation string
	Model     string
	Usage     float64
	Unit      string
	Status    string
}

type Recorder interface {
	RecordAIUsage(context.Context, Record) error
}
