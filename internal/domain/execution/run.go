package execution

import "fmt"

type Status string

const (
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

type Run struct {
	packagePath string
	target      string
	status      Status
	logs        []string
	duration    float64
}

func NewRun(packagePath, target string) Run {
	return Run{
		packagePath: packagePath,
		target:      target,
		status:      StatusRunning,
		logs:        []string{},
	}
}

func (run *Run) AppendLog(line string) {
	if line == "" {
		return
	}
	run.logs = append(run.logs, line)
}

func (run *Run) Complete(err error) {
	if err != nil {
		run.status = StatusFailed
		return
	}
	run.status = StatusSuccess
}

func (run Run) Status() Status      { return run.status }
func (run Run) PackagePath() string { return run.packagePath }
func (run Run) Target() string      { return run.target }
func (run Run) Logs() []string      { return append([]string(nil), run.logs...) }
func (run Run) Duration() float64   { return run.duration }

func (run Run) FormattedDuration() string {
	switch {
	case run.duration <= 0:
		return "0ms"
	case run.duration < 0.1:
		return fmt.Sprintf("%.0fms", run.duration*1000)
	default:
		if run.duration < 1 {
			return fmt.Sprintf("%.1fs", run.duration)
		}
		return fmt.Sprintf("%.2fs", run.duration)
	}
}

func (run *Run) AddDuration(sec float64) {
	if sec <= 0 {
		return
	}
	run.duration += sec
}
