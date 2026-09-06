package main

import (
	"testing"
	"time"
)

func TestMinimumPositiveDurationPreservesSubMillisecondWait(t *testing.T) {
	current := 4 * time.Millisecond
	wait := 500 * time.Microsecond

	if got := minimumPositiveDuration(current, wait); got != wait {
		t.Errorf("minimumPositiveDuration() = %v, want %v", got, wait)
	}
}

func TestGamepadRepeatUsesPressedKeyAfterPointerActivation(t *testing.T) {
	app := NewApp(DefaultConfig())
	kb := NewKeyboardState(LayoutQWERTY)
	rend := &Renderer{}
	pressed := findKeyPosition(t, func(key KeyDef) bool { return key.Code == KEY_Q })
	pointer := findKeyPosition(t, func(key KeyDef) bool { return key.ModifierType == "caps" })

	if !kb.SetCursor(pressed) {
		t.Fatal("could not position cursor on repeatable key")
	}
	app.handleAction(Action{Type: ActionPressStart}, kb, nil, rend)
	if app.repeatAction != ActionPressRepeat {
		t.Fatal("gamepad press did not start repeat")
	}
	if !pressPointerKey(kb, pointer, nil) || !kb.CapsActive {
		t.Fatal("pointer activation did not move cursor and enable caps lock")
	}

	app.handleAction(Action{Type: ActionPressRepeat}, kb, nil, rend)
	if !kb.CapsActive {
		t.Error("gamepad repeat activated pointer-selected key instead of pressed key")
	}
	if app.repeatPos != pressed {
		t.Errorf("repeat position = %v, want %v", app.repeatPos, pressed)
	}
}
