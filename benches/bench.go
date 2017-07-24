package benches

import (
	"fmt"
	"time"

	"github.com/estesp/bucketbench/driver"
)

// State represents the state of a benchmark object
type State int

// Type represents the type of benchmark
type Type int

// RunStatistics contains performance data from the benchmark run
// Each "step" from the benchmark is named and a map of the name
// to a millisecond duration for that step is provided
type RunStatistics struct {
	Durations map[string]int
	Errors    map[string]int
}

// Benchmark is the object form of a YAML-defined custom benchmark
// used to define the specific operations to perform
type Benchmark struct {
	Name     string
	Image    string
	Command  string //optionally override the default image CMD/ENTRYPOINT
	RootFs   string
	Detached bool
	Drivers  []DriverConfig
	Commands []string
}

// DriverConfig contains the YAML-defined parameters for running a
// benchmark against a specific driver type
type DriverConfig struct {
	Type       string
	Binary     string //optional path to specific client binary
	Threads    int
	Iterations int
}

// State constants
const (
	// Created represents a benchmark not yet run
	Created State = iota
	// Running represents a currently executing benchmark
	Running
	// Completed represents a finished benchmark run
	Completed
)

// Type constants
const (
	// Limit is a benchmark type for testing per-thread execution limits on the
	// hardware/environment
	Limit Type = iota
	// Custom is a YAML-defined series of container actions run as a benchmark
	Custom
)

// Bench is an interface to manage benchmark execution against a specific driver
type Bench interface {

	// Init initializes the benchmark (for example, verifies a daemon is running for daemon-centric
	// engines, pre-pulls images, etc.)
	Init(name string, driverType driver.Type, binaryPath, imageInfo, cmdOverride string, trace bool) error

	//Validates the any condition that need to be checked before actual banchmark run.
	//Helpful in testing operations required in benchmark for single run.
	Validate() error

	// Run executes the specified # of iterations against a specified # of
	// threads per benchmark against a specific engine driver type and collects
	// the statistics of each iteration and thread
	Run(threads, iterations int, commands []string) error

	// Stats returns the statistics of the benchmark run
	Stats() []RunStatistics

	// Elapsed returns the time.Duration that the benchmark took to execute
	Elapsed() time.Duration

	// State returns Created, Running, or Completed
	State() State

	// Type returns the type of benchmark
	Type() Type

	// Info returns a string with the driver type and custom benchmark name
	Info() string
}

// New creates an instance of the selected benchmark type
func New(btype Type) (Bench, error) {
	switch btype {
	case Limit:
		return &LimitBench{
			state: Created,
		}, nil
	case Custom:
		return &CustomBench{
			state: Created,
		}, nil
	default:
		return nil, fmt.Errorf("No such benchmark type: %v", btype)
	}
}
