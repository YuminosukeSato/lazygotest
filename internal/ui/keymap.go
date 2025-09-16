package ui

// KeyMap は lazgit 風のキーバインド（の一部）を定義する。
// 依存を減らすため Bubbles/key は使わず、単純なキー名の配列で表現する。
type KeyMap struct {
	Quit      []string
	FocusNext []string
	FocusPrev []string
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:      []string{"q", "ctrl+c"},
		FocusNext: []string{"tab"},
		FocusPrev: []string{"shift+tab"},
	}
}
