package service

const (
	DefaultAlpha = 0.3
)

type (
	// EMA is the Exponential Moving Average.
	EMA struct {
		alpha  float64
		value  EMAValue
		init   bool
		thread float64
	}

	EMAValue struct {
		CPUPercent    float64 `json:"cpu_percent"`
		MemoryPercent float64 `json:"memory_percent"`
		Connections   float64 `json:"connections"`
	}
)

// NewEMA creates an EMA.
func NewEMA(alpha, thread float64) *EMA {
	return &EMA{
		alpha:  alpha,
		thread: thread,
	}
}

// Update updates the EMA.
func (e *EMA) Update(value EMAValue) {
	if !e.init {
		e.value = value
		e.init = true
	} else {
		e.value.CPUPercent = e.alpha*value.CPUPercent + (1-e.alpha)*e.value.CPUPercent
		e.value.MemoryPercent = e.alpha*value.MemoryPercent + (1-e.alpha)*e.value.MemoryPercent
		e.value.Connections = e.alpha*value.Connections + (1-e.alpha)*e.value.Connections
	}
}

// Value returns the value of EMA.
func (e *EMA) Value() EMAValue {
	return e.value
}

// IsAnomaly returns if the value is anomaly.
func (e *EMA) IsAnomaly(value EMAValue) bool {
	return value.CPUPercent > e.value.CPUPercent*e.thread ||
		value.MemoryPercent > e.value.MemoryPercent*e.thread ||
		value.Connections > e.value.Connections*e.thread
}
