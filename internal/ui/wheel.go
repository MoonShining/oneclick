package ui

import (
	"image/color"
	"math"
	"voicepack/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type WheelItem struct {
	Item *config.AudioItem
}

// VoiceWheel 自定义轮盘控件
type VoiceWheel struct {
	widget.BaseWidget

	items   []WheelItem
	buttons []*DropWheelButton
	viewModel *config.ViewModel
	IsPopup bool // 是否为弹出模式（无边框、无背景）
}

func NewVoiceWheel(vm *config.ViewModel) *VoiceWheel {
	items := make([]WheelItem, 8)
	// 从配置加载已有的槽位
	for i := 0; i < 8; i++ {
		item := vm.GetChatWheelSlot(i)
		if item != nil {
			items[i].Item = vm.Config.GetAudioByID(item.ID)
		}
	}
	w := &VoiceWheel{items: items, viewModel: vm, IsPopup: false}
	w.ExtendBaseWidget(w)
	return w
}

// NewPopupVoiceWheel 创建弹出模式的轮盘（无边框、无背景）
func NewPopupVoiceWheel(vm *config.ViewModel) *VoiceWheel {
	w := NewVoiceWheel(vm)
	w.IsPopup = true
	return w
}

func (w *VoiceWheel) AtBtn(position *fyne.Position) *DropWheelButton {
	// 1. 获取轮盘自己的全局位置
	wheelPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(w)

	for _, btn := range w.buttons {
		// 2. 按钮位置是 相对于轮盘的
		// 所以 按钮全局坐标 = 轮盘全局坐标 + 按钮相对坐标
		btnGlobalX := wheelPos.X + btn.Position().X
		btnGlobalY := wheelPos.Y + btn.Position().Y

		// 3. 按钮大小
		w := btn.Size().Width
		h := btn.Size().Height

		// 4. 判断鼠标是否在矩形内
		xIn := position.X >= btnGlobalX && position.X <= btnGlobalX+w
		yIn := position.Y >= btnGlobalY && position.Y <= btnGlobalY+h

		if xIn && yIn {
			return btn
		}
	}
	return nil
}

func (w *VoiceWheel) CreateRenderer() fyne.WidgetRenderer {
	r := &wheelRenderer{wheel: w}
	r.Refresh()
	return r
}

type wheelRenderer struct {
	wheel   *VoiceWheel
	center  *canvas.Circle
	buttons []fyne.CanvasObject
}

func (r *wheelRenderer) Refresh() {
	r.buttons = nil
	r.wheel.buttons = nil
	size := r.wheel.Size()
	cx, cy := size.Width/2, size.Height/2
	radius := fyne.Min(cx, cy) * 0.6

	// 非弹出模式显示中心圆点
	if !r.wheel.IsPopup {
		r.center = canvas.NewCircle(color.NRGBA{0x40, 0x80, 0xFF, 0xFF})
		r.center.Resize(fyne.NewSize(40, 40))
		r.center.Move(fyne.NewPos(cx-20, cy-20))
		r.buttons = append(r.buttons, r.center)
	}

	// 环形分布按钮
	count := len(r.wheel.items)
	btnWidth := float32(150)
	btnHeight := float32(30)

	for i, item := range r.wheel.items {
		angle := 2*math.Pi*float64(i)/float64(count) - math.Pi/2
		x := cx + float32(math.Cos(angle))*radius - btnWidth/2
		y := cy + float32(math.Sin(angle))*radius - btnHeight/2

		idx := i
		text := "[空]"
		if item.Item != nil {
			text = item.Item.Name
		}

		// 根据模式创建不同按钮
		var btn *DropWheelButton
		if r.wheel.IsPopup {
			// 弹出模式使用半透明背景按钮
			btn = NewDropWheelButton(text, i, nil)
		} else {
			// 普通模式使用带边框按钮
			dwb := NewDropWheelButton(text, i, func(item *config.AudioItem) {
				r.wheel.items[idx].Item = item
				if r.wheel.viewModel != nil {
					r.wheel.viewModel.SetChatWheelSlot(idx, item)
				}
			})
			btn = dwb
		}
		btn.Resize(fyne.NewSize(btnWidth, btnHeight))
		btn.Move(fyne.NewPos(x, y))
		r.buttons = append(r.buttons, btn)
		r.wheel.buttons = append(r.wheel.buttons, btn)
	}
}

func (r *wheelRenderer) Layout(size fyne.Size) {
	// 确保 wheel 使用正确的尺寸
	r.wheel.BaseWidget.Resize(size)
	// 调用 Refresh 重新布局所有元素
	r.Refresh()
}

func (r *wheelRenderer) MinSize() fyne.Size {
	return fyne.NewSize(400, 400)
}

func (r *wheelRenderer) Objects() []fyne.CanvasObject {
	return r.buttons
}

func (r *wheelRenderer) Destroy() {}
