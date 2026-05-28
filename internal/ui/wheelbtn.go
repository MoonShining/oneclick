package ui

import (
	"image/color"
	"voicepack/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// 可接收拖放的轮盘按钮（你的样式 + 感知拖拽）
type DropWheelButton struct {
	widget.BaseWidget
	Text   string
	Index  int                          // 按钮在轮盘的位置
	OnDrop func(item *config.AudioItem) // 拖放回调

	// 缓存的渲染元素
	label  *widget.Label
	border *canvas.Rectangle
}

func NewDropWheelButton(text string, index int, onDrop func(item *config.AudioItem)) *DropWheelButton {
	d := &DropWheelButton{
		Text:   text,
		Index:  index,
		OnDrop: onDrop,
	}
	d.ExtendBaseWidget(d)
	return d
}

func (d *DropWheelButton) CreateRenderer() fyne.WidgetRenderer {
	d.label = widget.NewLabel(d.Text)
	d.label.Truncation = fyne.TextTruncateClip
	d.label.Alignment = fyne.TextAlignCenter

	d.border = canvas.NewRectangle(color.Transparent)
	d.border.StrokeColor = color.Black
	d.border.StrokeWidth = 1.0

	content := container.NewStack(d.border, d.label)
	return widget.NewSimpleRenderer(content)
}

func (d *DropWheelButton) Refresh() {
	// 更新 label 文本
	if d.label != nil {
		d.label.SetText(d.Text)
	}
	d.BaseWidget.Refresh()
}

func (d *DropWheelButton) Dropped(item *config.AudioItem) {
	d.Text = item.Name
	if d.OnDrop != nil {
		d.OnDrop(item)
	}
	d.Refresh()
}
