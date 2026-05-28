package ui

import (
	"os"
	"path/filepath"
	"strings"

	"voicepack/internal/audio"
	"voicepack/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type MainWindow struct {
	window  fyne.Window
	viewModel *config.ViewModel
	config  *config.Config // keep for compatibility during refactor
	player  *audio.Player
	onClose func()
	app     fyne.App

	// UI components
	tree         *widget.Tree
	audioList    *widget.List
	volumeSlider *widget.Slider
	hotkeyEntry  *HotKeyEntry
	popupWindow  fyne.Window
	popupWheel   *VoiceWheel
	wheel        *VoiceWheel

	// State
	selectedGroup *config.Group
	selectedAudio *config.AudioItem
	draggedAudio  *config.AudioItem
}

func NewMainWindow(app fyne.App, vm *config.ViewModel, p *audio.Player, onClose func()) *MainWindow {
	w := &MainWindow{
		window:  app.NewWindow("OneClick 语音包"),
		viewModel: vm,
		config:  vm.Config,
		player:  p,
		onClose: onClose,
		app:     app,
	}

	// Connect player volume to view model binding - when volume changes in config, update player
	w.viewModel.Volume.AddListener(binding.NewDataListener(func() {
		value, _ := w.viewModel.Volume.Get()
		w.player.SetVolume(value)
	}))

	w.buildUI()
	w.setupKeyListener()
	w.selectFirstGroup()

	return w
}

func (w *MainWindow) buildUI() {
	left := w.buildLeftPanel()
	right := w.buildRightPanel()
	middle := w.buildMiddlePanel()

	rightSplit := container.NewHSplit(middle, right)
	rightSplit.Offset = 0.7

	mainSplit := container.NewHSplit(left, rightSplit)
	mainSplit.Offset = 0.20

	w.window.SetContent(mainSplit)
	w.window.Resize(fyne.NewSize(1000, 700))
}

func (w *MainWindow) buildLeftPanel() fyne.CanvasObject {
	w.tree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			// 返回给定节点的子节点
			var children []widget.TreeNodeID
			if id == w.tree.Root {
				// 根节点返回所有分组作为第一级列表
				for _, group := range w.config.Groups {
					children = append(children, widget.TreeNodeID(group.ID))
				}
			}
			// 当前不支持分组嵌套，所以非根节点没有子节点
			return children
		},
		func(id widget.TreeNodeID) bool {
			// 判断节点是否有子节点（是分支）
			// 只有根节点有子节点，分组本身不嵌套
			return id == w.tree.Root
		},
		func(isBranch bool) fyne.CanvasObject {
			return widget.NewLabel("Group")
		},
		func(id widget.TreeNodeID, isBranch bool, node fyne.CanvasObject) {
			group := w.config.GetGroupByID(string(id))
			if group == nil {
				return
			}
			node.(*widget.Label).SetText(group.Name)
		},
	)

	w.tree.Root = ""

	w.tree.OnSelected = func(id widget.TreeNodeID) {
		group := w.config.GetGroupByID(string(id))
		if group != nil {
			w.selectedGroup = group
			w.audioList.Refresh()
		}
	}

	return container.NewPadded(w.tree)
}

func (w *MainWindow) buildMiddlePanel() fyne.CanvasObject {
	w.audioList = widget.NewList(
		func() int {
			// 修复：应该检查 selectedGroup 而不是 selectedAudio
			if w.selectedGroup == nil {
				return 0
			}
			items := w.config.GetAudioByGroup(w.selectedGroup.ID)
			return len(items)
		},
		func() fyne.CanvasObject {
			return NewListLabel(w.wheel, "template", nil, func(item *config.AudioItem) {
				w.playAudio(item)
			})
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*ListLabel)
			items := w.config.GetAudioByGroup(w.selectedGroup.ID)
			item := items[i]
			label.SetText(item.Name)
			label.item = item
		},
	)

	label := widget.NewLabel("拖动到右侧轮盘添加音效")
	return container.NewBorder(label, nil, nil, nil, w.audioList)
}

func (w *MainWindow) buildRightPanel() fyne.CanvasObject {
	title := widget.NewLabel("语音轮盘快捷键")
	title.TextStyle = fyne.TextStyle{Bold: true}

	w.hotkeyEntry = NewHotKeyEntry()
	w.hotkeyEntry.Bind(w.viewModel.ChatWheelHotkey)
	w.hotkeyEntry.OnChanged = func(s string) {
		if len(s) > 1 {
			s = s[:1]
			w.hotkeyEntry.SetText(s)
		}
		// Auto-save handled by view model binding
	}
	hotkeyBox := container.NewHBox(title, w.hotkeyEntry)

	w.wheel = NewVoiceWheel(w.viewModel)
	gridBox := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(w.wheel),
		layout.NewSpacer(),
	)

	// Volume is 0-100 in config, directly matches slider range
	w.volumeSlider = widget.NewSliderWithData(0, 100, w.viewModel.Volume)
	volumeBox := container.NewBorder(nil, nil, widget.NewLabel("音量:"), nil, w.volumeSlider)

	content := container.NewBorder(
		container.NewVBox(hotkeyBox, widget.NewSeparator()),
		container.NewVBox(volumeBox),
		nil,
		nil,
		gridBox)

	return content
}

func (w *MainWindow) playAudio(item *config.AudioItem) {
	path := item.Path
	if !filepath.IsAbs(path) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			home, err := os.UserHomeDir()
			if err == nil {
				path = filepath.Join(home, ".oneclick", "audio", item.Path)
			}
		}
	}

	var loaded *audio.Audio
	var err error

	ext := getExtension(path)
	switch strings.ToLower(ext) {
	case "mp3":
		loaded, err = w.player.LoadMP3(path)
	case "wav":
		loaded, err = w.player.LoadWAV(path)
	default:
		loaded, err = w.player.LoadMP3(path)
	}

	if err != nil {
		dialog.ShowError(err, w.window)
		return
	}

	err = w.player.Play(loaded)
	if err != nil {
		dialog.ShowError(err, w.window)
		return
	}
}

func (w *MainWindow) Show() {
	w.window.Show()
}

func (w *MainWindow) Hide() {
	w.window.Hide()
}

func (w *MainWindow) Close() {
	w.player.Stop()
	w.window.Close()
	if w.onClose != nil {
		w.onClose()
	}
}

func (w *MainWindow) selectFirstGroup() {
	if len(w.config.Groups) > 0 {
		firstGroup := w.config.Groups[0]
		w.tree.OpenBranch(w.tree.Root)
		id := widget.TreeNodeID(firstGroup.ID)
		w.tree.Select(id)
		w.tree.ScrollTo(id)

		audioItems := w.config.GetAudioByGroup(firstGroup.ID)
		if len(audioItems) > 0 {
			w.selectedAudio = audioItems[0]
			w.draggedAudio = audioItems[0]
		}
	}
}

func (w *MainWindow) setupKeyListener() {
	// 监听键盘按下和释放事件
	convertKey := func(e *fyne.KeyEvent) string {
		keyStr := string(e.Name)
		if len(keyStr) == 0 || keyStr[0] < ' ' {
			// 处理特殊按键
			switch e.Name {
			case fyne.KeySpace:
				keyStr = " "
			case fyne.KeyEnter:
				keyStr = "\n"
			case fyne.KeyTab:
				keyStr = "\t"
			}
		}
		return strings.ToLower(keyStr)
	}

	c := w.window.Canvas().(desktop.Canvas)
	c.SetOnKeyDown(func(e *fyne.KeyEvent) {
		wheel := w.config.GetActiveChatWheel()
		if wheel == nil || wheel.Hotkey == "" {
			return
		}

		// 将按键转换为字符串进行比较
		keyStr := convertKey(e)
		if keyStr == wheel.Hotkey {
			w.showPopupChatWheel()
		}
	})

	c.SetOnKeyUp(func(e *fyne.KeyEvent) {
		wheel := w.config.GetActiveChatWheel()
		if wheel == nil || wheel.Hotkey == "" {
			return
		}
		// 将按键转换为字符串进行比较
		keyStr := convertKey(e)
		if keyStr == wheel.Hotkey {
			w.hidePopupChatWheel()
		}
	})
}

func (w *MainWindow) showPopupChatWheel() {
	if w.popupWindow != nil {
		// 如果窗口已存在，确保它是可见的
		w.popupWindow.Show()
		return
	}

	// 创建弹出模式的轮盘
	w.popupWheel = NewPopupVoiceWheel(w.viewModel)
	w.popupWheel.Resize(fyne.NewSize(500, 500))

	// 创建独立的无边框窗口
	w.popupWindow = w.app.NewWindow("popup wheel")
	// 设置窗口无边框
	w.popupWindow.SetPadded(false)
	w.popupWindow.SetFixedSize(true)
	w.popupWindow.Resize(fyne.NewSize(500, 500))

	// 获取屏幕中心位置
	w.popupWindow.CenterOnScreen()
	w.popupWindow.SetContent(w.popupWheel)

	// 显示窗口
	w.popupWindow.Show()
}

func (w *MainWindow) hidePopupChatWheel() {
	if w.popupWindow != nil {
		w.popupWindow.Hide()
	}
}

func getExtension(path string) string {
	return strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
}