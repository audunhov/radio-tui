package tui

import (
	"fmt"
	"radio-tui/api"
	"radio-tui/models"
	"radio-tui/player"
	"radio-tui/storage"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type focusState int

const (
	focusInput focusState = iota
	focusSearchList
	focusFavList
)

const (
	// Layout constants
	windowWidthBreakpoint = 80 // Width at which we switch to single-column layout
	paneHeightOverhead    = 8  // Combined height of input (3), status (3), help (1), and spacing (1)

	// Border offsets for list sizing
	verticalBorderOffset   = 4 // Total height taken by list borders in single-column mode (2 lists * 2 border lines)
	horizontalBorderOffset = 2 // Total height taken by list borders in two-column mode (1 list height * 2 border lines)
	sidePaddingOffset      = 2 // Width taken by list borders (Left + Right)
)

var (
	activeBorderColor   = lipgloss.Color("62")  // Purple
	inactiveBorderColor = lipgloss.Color("240") // Dark gray

	activeBorderStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(activeBorderColor)
	inactiveBorderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(inactiveBorderColor)

	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(2)
)

type Model struct {
	focus       focusState
	searchList  list.Model
	favList     list.Model
	input       textinput.Model
	player      *player.AudioService
	api         *api.Client
	storage     *storage.StorageService
	favorites   []models.Station
	playing     *models.Station
	currentSong string
	err         error
	width       int
	height      int
}

type item struct {
	station models.Station
}

func (i item) Title() string       { return i.station.Name }
func (i item) Description() string { return i.station.Country + " | " + i.station.Codec }
func (i item) FilterValue() string { return i.station.Name }

func NewModel(p *player.AudioService, s *storage.StorageService) Model {
	ti := textinput.New()
	ti.Placeholder = "Search radio stations..."
	ti.Prompt = "🔍 "
	ti.Focus()

	sl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	sl.Title = "Search Results"
	sl.SetShowFilter(false)
	sl.SetShowHelp(false)
	sl.SetShowPagination(false)

	fl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	fl.Title = "Favorites"
	fl.SetShowFilter(false)
	fl.SetShowHelp(false)
	fl.SetShowPagination(false)

	favs, _ := s.LoadFavorites()
	var favItems []list.Item
	for _, f := range favs {
		favItems = append(favItems, item{station: f})
	}
	fl.SetItems(favItems)

	return Model{
		focus:      focusInput,
		searchList: sl,
		favList:    fl,
		input:      ti,
		player:     p,
		api:        api.NewClient(),
		storage:    s,
		favorites:  favs,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.waitForMetadata())
}

type searchMsg []models.Station
type metadataMsg string
type errMsg error

func (m Model) waitForMetadata() tea.Cmd {
	return func() tea.Msg {
		return metadataMsg(<-m.player.Updates)
	}
}

func (m Model) search(query string) tea.Cmd {
	return func() tea.Msg {
		stations, err := m.api.Search(query)
		if err != nil {
			return errMsg(err)
		}
		return searchMsg(stations)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowSize(msg)
	case tea.KeyMsg:
		if model, cmd, done := m.handleKeys(msg); done {
			return model, cmd
		}
	case searchMsg:
		var items []list.Item
		for _, s := range msg {
			items = append(items, item{station: s})
		}
		m.searchList.SetItems(items)
	case metadataMsg:
		m.currentSong = string(msg)
		return m, m.waitForMetadata()
	case errMsg:
		m.err = msg
	}

	return m.updateFocusedComponent(msg)
}

func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height

	var leftWidth, rightWidth, searchHeight, favHeight int
	if msg.Width < windowWidthBreakpoint {
		leftWidth = msg.Width - sidePaddingOffset
		rightWidth = msg.Width - sidePaddingOffset
		availableHeight := max(msg.Height-paneHeightOverhead-verticalBorderOffset, 0)
		searchHeight = availableHeight / 2
		favHeight = availableHeight - searchHeight
	} else {
		leftWidth = (msg.Width / 2) - sidePaddingOffset
		rightWidth = msg.Width - (msg.Width / 2) - sidePaddingOffset
		searchHeight = max(msg.Height-paneHeightOverhead-horizontalBorderOffset, 0)
		favHeight = searchHeight
	}

	m.searchList.SetSize(leftWidth, searchHeight)
	m.favList.SetSize(rightWidth, favHeight)
}

func (m *Model) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+c":
		m.player.Stop()
		return m, tea.Quit, true
	case "q":
		if m.focus != focusInput {
			m.player.Stop()
			return m, tea.Quit, true
		}
	case "tab":
		m.cycleFocus(true)
		return m, nil, true
	case "shift+tab":
		m.cycleFocus(false)
		return m, nil, true
	case "esc":
		if m.focus != focusInput {
			m.focus = focusInput
			m.input.Focus()
		}
		return m, nil, true
	case "enter", " ":
		return m.handleActionKey(msg)
	case "f":
		if m.focus != focusInput {
			m.handleToggleFavorite()
		}
		return m, nil, true
	}
	return m, nil, false
}

func (m *Model) cycleFocus(forward bool) {
	if forward {
		switch m.focus {
		case focusInput:
			m.focus = focusSearchList
			m.input.Blur()
		case focusSearchList:
			m.focus = focusFavList
		default:
			m.focus = focusInput
			m.input.Focus()
		}
	} else {
		switch m.focus {
		case focusInput:
			m.focus = focusFavList
			m.input.Blur()
		case focusFavList:
			m.focus = focusSearchList
		default:
			m.focus = focusInput
			m.input.Focus()
		}
	}
}

func (m *Model) handleActionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if m.focus == focusInput {
		if msg.String() == "enter" {
			m.focus = focusSearchList
			m.input.Blur()
			return m, m.search(m.input.Value()), true
		}
		return m, nil, false
	}

	var selected models.Station
	switch m.focus {
	case focusSearchList:
		if i, ok := m.searchList.SelectedItem().(item); ok {
			selected = i.station
		}
	case focusFavList:
		if i, ok := m.favList.SelectedItem().(item); ok {
			selected = i.station
		}
	}

	if selected.ID != "" {
		if m.playing != nil && m.playing.ID == selected.ID {
			m.player.TogglePause()
		} else if selected.URL != "" {
			m.playing = &selected
			m.currentSong = "Loading..."
			m.player.Play(selected.URL)
		}
	}
	return m, nil, true
}

func (m *Model) handleToggleFavorite() {
	var selected models.Station
	switch m.focus {
	case focusSearchList:
		if i, ok := m.searchList.SelectedItem().(item); ok {
			selected = i.station
		}
	case focusFavList:
		if i, ok := m.favList.SelectedItem().(item); ok {
			selected = i.station
		}
	}

	if selected.ID != "" {
		m.toggleFavorite(selected)
	}
}

func (m *Model) updateFocusedComponent(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focus {
	case focusInput:
		m.input, cmd = m.input.Update(msg)
	case focusSearchList:
		m.searchList, cmd = m.searchList.Update(msg)
	case focusFavList:
		m.favList, cmd = m.favList.Update(msg)
	}
	return m, cmd
}

func (m *Model) toggleFavorite(s models.Station) {
	found := -1
	for i, f := range m.favorites {
		if f.ID == s.ID {
			found = i
			break
		}
	}

	if found != -1 {
		m.favorites = append(m.favorites[:found], m.favorites[found+1:]...)
	} else {
		m.favorites = append(m.favorites, s)
	}

	m.storage.SaveFavorites(m.favorites)

	var favItems []list.Item
	for _, f := range m.favorites {
		favItems = append(favItems, item{station: f})
	}
	m.favList.SetItems(favItems)
}

func (m Model) getHelpText() string {
	switch m.focus {
	case focusInput:
		return "enter: search • tab: navigate • ctrl+c: quit"
	case focusSearchList:
		return "enter/space: play/pause • f: favorite • tab: navigate • esc: focus search • q: quit"
	case focusFavList:
		return "enter/space: play/pause • f: unfavorite • tab: navigate • esc: focus search • q: quit"
	}
	return ""
}

func (m Model) View() string {
	inputStyle := inactiveBorderStyle
	if m.focus == focusInput {
		inputStyle = activeBorderStyle
	}

	searchStyle := inactiveBorderStyle
	if m.focus == focusSearchList {
		searchStyle = activeBorderStyle
	}

	favStyle := inactiveBorderStyle
	if m.focus == focusFavList {
		favStyle = activeBorderStyle
	}

	inputView := inputStyle.Width(m.width - sidePaddingOffset).Render(m.input.View())

	var leftWidth, rightWidth, searchHeight, favHeight int
	if m.width < windowWidthBreakpoint {
		leftWidth = m.width - sidePaddingOffset
		rightWidth = m.width - sidePaddingOffset
		availableHeight := max(m.height-paneHeightOverhead-verticalBorderOffset, 0)
		searchHeight = availableHeight / 2
		favHeight = availableHeight - searchHeight
	} else {
		leftWidth = (m.width / 2) - sidePaddingOffset
		rightWidth = m.width - (m.width / 2) - sidePaddingOffset
		searchHeight = max(m.height-paneHeightOverhead-horizontalBorderOffset, 0)
		favHeight = searchHeight
	}

	searchView := searchStyle.Width(leftWidth).Height(searchHeight).Render(m.searchList.View())
	favView := favStyle.Width(rightWidth).Height(favHeight).Render(m.favList.View())

	var listsView string
	if m.width < windowWidthBreakpoint {
		listsView = lipgloss.JoinVertical(lipgloss.Left, searchView, favView)
	} else {
		listsView = lipgloss.JoinHorizontal(lipgloss.Top, searchView, favView)
	}

	statusText := "Not playing"
	if m.playing != nil {
		state := "Playing"
		if m.player.IsPaused() {
			state = "Paused"
		}
		statusText = fmt.Sprintf("[%s] %s | %s", state, m.playing.Name, m.currentSong)
	}
	if m.err != nil {
		statusText = fmt.Sprintf("Error: %v", m.err)
	}

	statusView := inactiveBorderStyle.Width(m.width - sidePaddingOffset).Render(statusText)
	helpView := helpStyle.Render(m.getHelpText())

	return lipgloss.JoinVertical(lipgloss.Left, inputView, listsView, statusView, helpView)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
