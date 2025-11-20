//go:build tview

package ui

import "testing"

func TestDefaultTheme(t *testing.T) {
	t.Parallel()
	theme := DefaultTheme()

	if theme.BGPrimary == 0 {
		t.Error("BGPrimary should be set")
	}
	if theme.TextPrimary == 0 {
		t.Error("TextPrimary should be set")
	}
	if theme.BorderFocus == 0 {
		t.Error("BorderFocus should be set")
	}
	if theme.Success == 0 {
		t.Error("Success should be set")
	}
}

func TestStatusIcons(t *testing.T) {
	t.Parallel()
	if IconPass == "" {
		t.Error("IconPass should not be empty")
	}
	if IconFail == "" {
		t.Error("IconFail should not be empty")
	}
	if IconRunning == "" {
		t.Error("IconRunning should not be empty")
	}
	if IconPending == "" {
		t.Error("IconPending should not be empty")
	}
}

func TestLogLevels(t *testing.T) {
	t.Parallel()
	if LogInfo == "" {
		t.Error("LogInfo should not be empty")
	}
	if LogWarn == "" {
		t.Error("LogWarn should not be empty")
	}
	if LogError == "" {
		t.Error("LogError should not be empty")
	}
	if LogDebug == "" {
		t.Error("LogDebug should not be empty")
	}
}
