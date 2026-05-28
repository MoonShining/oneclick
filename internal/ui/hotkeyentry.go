package ui

import (
	"fyne.io/fyne/v2/widget"
)

// 获得焦点时自动清空的 Entry
type HotKeyEntry struct {
	widget.Entry
}

func NewHotKeyEntry() *HotKeyEntry {
	entry := &HotKeyEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

// 👇 核心：获得焦点时清空
func (e *HotKeyEntry) FocusGained() {
	// 调用原生逻辑
	e.Entry.FocusGained()
	// 清空内容
	e.SetText("")
}
