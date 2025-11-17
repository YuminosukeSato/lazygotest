package ui

// WithFilterUI injects a modal/input handler that receives ApplyFilter callback.
func WithFilterUI(fn func(apply func(string))) Option {
	return func(a *App) {
		a.filterUI = fn
	}
}

// WithHelpUI injects a help overlay handler.
func WithHelpUI(fn func()) Option {
	return func(a *App) {
		a.helpUI = fn
	}
}
