package execution

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

func NewRun(pkg, target string) Run {
	return Run{
		packagePath: pkg,
		target:      target,
		status:      StatusRunning,
		logs:        []string{},
	}
}

func (r *Run) AppendLog(line string) {
	if line == "" {
		return
	}
	r.logs = append(r.logs, line)
}

func (r *Run) Complete(err error) {
	if err != nil {
		r.status = StatusFailed
		return
	}
	r.status = StatusSuccess
}

func (r Run) Status() Status      { return r.status }
func (r Run) PackagePath() string { return r.packagePath }
func (r Run) Target() string      { return r.target }
func (r Run) Logs() []string      { return append([]string(nil), r.logs...) }
func (r Run) Duration() float64   { return r.duration }

func (r *Run) AddDuration(sec float64) {
	if sec <= 0 {
		return
	}
	r.duration += sec
}
