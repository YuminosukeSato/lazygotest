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
	case run.duration < 1:
		return formatMillis(run.duration)
	default:
		return formatSeconds(run.duration)
	}
}

func formatMillis(sec float64) string {
	ms := sec * 1000
	return formatFloat(ms, 0, "ms")
}

func formatSeconds(sec float64) string {
	return formatFloat(sec, 2, "s")
}

func formatFloat(val float64, prec int, suffix string) string {
	return trimTrailingZeros(val, prec) + suffix
}

func trimTrailingZeros(val float64, prec int) string {
	// simple formatting without importing fmt to keep deps minimal
	pow := 1.0
	for i := 0; i < prec; i++ {
		pow *= 10
	}
	rounded := float64(int(val*pow+0.5)) / pow
	str := floatToString(rounded, prec)
	for len(str) > 0 && str[len(str)-1] == '0' {
		str = str[:len(str)-1]
	}
	if len(str) > 0 && str[len(str)-1] == '.' {
		str = str[:len(str)-1]
	}
	return str
}

func floatToString(val float64, prec int) string {
	// minimal fmt.Sprintf replacement for fixed precision
	// note: not handling NaN/Inf as durations won't carry them here
	intPart := int(val)
	scale := float64(pow10(prec))
	frac := int((val - float64(intPart)) * scale)
	if prec == 0 {
		return itoa(intPart)
	}
	return itoa(intPart) + "." + padLeft(itoa(frac), prec)
}

func pow10(n int) int {
	p := 1
	for i := 0; i < n; i++ {
		p *= 10
	}
	return p
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	buf := make([]byte, width)
	for i := 0; i < width-len(s); i++ {
		buf[i] = '0'
	}
	copy(buf[width-len(s):], s)
	return string(buf)
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	var buf [16]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return sign + string(buf[i:])
}
func (run *Run) AddDuration(sec float64) {
	if sec <= 0 {
		return
	}
	run.duration += sec
}
