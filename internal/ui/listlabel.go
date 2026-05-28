package ui

import (
	"voicepack/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type ListLabel struct {
	widget.Label
	item *config.AudioItem // 这里放你的音频路径、ID、任意信息

	wheel *VoiceWheel
	pos   *fyne.Position

	onDoubleTap func(item *config.AudioItem)
}

func NewListLabel(wheel *VoiceWheel, text string, item *config.AudioItem, onDoubleTap func(item *config.AudioItem)) *ListLabel {
	l := &ListLabel{}
	l.ExtendBaseWidget(l)
	l.SetText(text)
	l.item = item
	l.wheel = wheel
	l.onDoubleTap = onDoubleTap
	return l
}

// 👇 双击事件（核心）
func (dl *ListLabel) DoubleTapped(e *fyne.PointEvent) {
	dl.onDoubleTap(dl.item)
}

// 实现 Draggable：拖动中
func (dl *ListLabel) Dragged(e *fyne.DragEvent) {
	if dl.pos == nil {
		dl.pos = &e.AbsolutePosition
		return
	}

	dl.pos = &e.AbsolutePosition
}

// 实现 Draggable：拖动结束
func (dl *ListLabel) DragEnd() {
	btn := dl.wheel.AtBtn(dl.pos)
	dl.pos = nil

	if btn != nil {
		btn.Dropped(dl.item)
	}
}
