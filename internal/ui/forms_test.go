package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/assert"
)

func TestArrowModel_KeyMessages(t *testing.T) {
	// Create a minimal huh Form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Value(new(string)),
		),
	)

	m := &arrowModel{
		form: form,
	}

	// 1. Test Init
	m.Init()

	// 2. Test Left Arrow KeyMsg (should set backPressed and exit)
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	typedModel, ok := newModel.(*arrowModel)
	assert.True(t, ok)
	assert.True(t, typedModel.backPressed)
	assert.NotNil(t, cmd)

	// 3. Test Right Arrow KeyMsg (should translate key to Enter and not crash)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})

	// 4. Test View (should render successfully)
	viewStr := m.View()
	assert.NotEmpty(t, viewStr)
}
