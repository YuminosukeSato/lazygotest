//go:build tview

package ui

import "github.com/gdamore/tcell/v2"

type Theme struct {
	BGPrimary        tcell.Color
	BGSecondary      tcell.Color
	TextPrimary      tcell.Color
	TextSecondary    tcell.Color
	BorderNormal     tcell.Color
	BorderFocus      tcell.Color
	Success          tcell.Color
	Failure          tcell.Color
	Running          tcell.Color
	Info             tcell.Color
	Warning          tcell.Color
	Error            tcell.Color
	TitleText        tcell.Color
	TitleBG          tcell.Color
	StatusBarText    tcell.Color
	StatusBarBG      tcell.Color
	StatusBarKeyBG   tcell.Color
	StatusBarKeyText tcell.Color
}

func DefaultTheme() Theme {
	return Theme{
		BGPrimary:        tcell.NewRGBColor(30, 30, 46),
		BGSecondary:      tcell.NewRGBColor(49, 50, 68),
		TextPrimary:      tcell.NewRGBColor(205, 214, 244),
		TextSecondary:    tcell.NewRGBColor(186, 194, 222),
		BorderNormal:     tcell.NewRGBColor(69, 71, 90),
		BorderFocus:      tcell.NewRGBColor(137, 180, 250),
		Success:          tcell.NewRGBColor(166, 227, 161),
		Failure:          tcell.NewRGBColor(243, 139, 168),
		Running:          tcell.NewRGBColor(249, 226, 175),
		Info:             tcell.NewRGBColor(137, 180, 250),
		Warning:          tcell.NewRGBColor(249, 226, 175),
		Error:            tcell.NewRGBColor(243, 139, 168),
		TitleText:        tcell.NewRGBColor(205, 214, 244),
		TitleBG:          tcell.NewRGBColor(49, 50, 68),
		StatusBarText:    tcell.NewRGBColor(205, 214, 244),
		StatusBarBG:      tcell.NewRGBColor(49, 50, 68),
		StatusBarKeyBG:   tcell.NewRGBColor(137, 180, 250),
		StatusBarKeyText: tcell.NewRGBColor(30, 30, 46),
	}
}

type StatusIcon string

const (
	IconPass    StatusIcon = "✓"
	IconFail    StatusIcon = "✗"
	IconRunning StatusIcon = "⏳"
	IconPending StatusIcon = "○"
	IconSkip    StatusIcon = "⊘"
)

type LogLevel string

const (
	LogInfo  LogLevel = "[INFO]"
	LogWarn  LogLevel = "[WARN]"
	LogError LogLevel = "[ERROR]"
	LogDebug LogLevel = "[DEBUG]"
)
