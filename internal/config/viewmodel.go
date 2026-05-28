package config

import (
	"fyne.io/fyne/v2/data/binding"
)

// ViewModel is the MVVM view model for Config, providing bound properties
// that automatically save to disk when any change occurs.
type ViewModel struct {
	Config *Config

	// Bound properties that two-way bind with UI components
	// Volume is stored as 0-100 directly matching slider range
	Volume         binding.Float
	ChatWheelHotkey binding.String
	// Slots are handled specially since it's a slice
}

// NewViewModel creates a new ViewModel wrapping the given config.
// It automatically sets up bindings and triggers auto-save on changes.
func NewViewModel(cfg *Config) *ViewModel {
	vm := &ViewModel{
		Config:         cfg,
		Volume:         binding.BindFloat(&cfg.Volume), // 0-100 directly
		ChatWheelHotkey: binding.NewString(),
	}

	// Initialize hotkey from config
	wheel := cfg.GetActiveChatWheel()
	if wheel != nil {
		vm.ChatWheelHotkey.Set(wheel.Hotkey)
	}

	// Add change listeners that auto-save and sync to underlying config
	vm.Volume.AddListener(binding.NewDataListener(func() {
		// Volume is already synced via BindFloat, just save
		vm.Config.Save()
	}))

	vm.ChatWheelHotkey.AddListener(binding.NewDataListener(func() {
		// Sync hotkey to config then save
		wheel := vm.Config.GetActiveChatWheel()
		if wheel != nil {
			hotkey, _ := vm.ChatWheelHotkey.Get()
			wheel.Hotkey = hotkey
			vm.Config.Save()
		}
	}))

	return vm
}

// GetChatWheelSlot gets the slot at the given index
func (vm *ViewModel) GetChatWheelSlot(index int) *AudioItem {
	wheel := vm.Config.GetActiveChatWheel()
	if wheel == nil || index < 0 || index >= len(wheel.Slots) {
		return nil
	}
	return wheel.Slots[index]
}

// SetChatWheelSlot sets the slot at the given index and triggers auto-save
func (vm *ViewModel) SetChatWheelSlot(index int, item *AudioItem) {
	wheel := vm.Config.GetActiveChatWheel()
	if wheel == nil || index < 0 || index >= len(wheel.Slots) {
		return
	}
	wheel.Slots[index] = item
	vm.Config.Save()
}

// SlotsCount returns the number of chat wheel slots
func (vm *ViewModel) SlotsCount() int {
	wheel := vm.Config.GetActiveChatWheel()
	if wheel == nil {
		return 0
	}
	return len(wheel.Slots)
}
