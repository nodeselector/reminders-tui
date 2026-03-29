package model

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nodeselector/reminders-tui/internal/remindctl"
	"github.com/nodeselector/reminders-tui/internal/ui"
)

// View represents which view is active.
type View int

const (
	ViewToday View = iota
	ViewUpcoming
	ViewLists
	ViewAll
)

func (v View) String() string {
	switch v {
	case ViewToday:
		return "Today"
	case ViewUpcoming:
		return "Upcoming"
	case ViewLists:
		return "Lists"
	case ViewAll:
		return "All"
	default:
		return "?"
	}
}

// Mode represents the current interaction mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeLoading
	ModeSearch
	ModeAdd
	ModeAddField    // cycling through optional add fields
	ModeEdit
	ModeEditField
	ModeConfirmDelete
	ModeHelp
)

// displayItem is a row in the rendered list -- either a group header or a reminder.
type displayItem struct {
	isHeader bool
	header   string
	overdue  bool // for group header styling
	reminder remindctl.Reminder
	index    int // index into filtered reminders (only for non-headers)
}

// Model is the top-level Bubble Tea model.
type Model struct {
	client *remindctl.Client

	// Data
	reminders []remindctl.Reminder
	lists     []remindctl.ReminderList

	// UI state
	view         View
	mode         Mode
	cursor       int           // cursor position in displayItems
	displayItems []displayItem // computed view items
	width        int
	height       int

	// Search
	searchInput textinput.Model
	searchQuery string

	// Add mode
	addInput    textinput.Model
	addTitle    string
	addList     string
	addDue      string
	addPriority string
	addStep     int // 0=title, 1=list, 2=due, 3=priority

	// Edit mode
	editInput   textinput.Model
	editID      string
	editStep    int // 0=title, 1=due, 2=priority
	editTitle   string
	editDue     string
	editPriority string

	// Status
	statusMsg   string
	statusErr   bool
	statusTimer int // ticks remaining

	// Loading
	spinner spinner.Model
	loading bool

	// Confirm delete
	deleteID    string
	deleteTitle string
}

// Messages
type (
	remindersLoadedMsg struct {
		reminders []remindctl.Reminder
		lists     []remindctl.ReminderList
		err       error
	}
	actionDoneMsg struct {
		msg string
		err error
	}
	statusTickMsg struct{}
)

// New creates a new model.
func New(client *remindctl.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ui.SpinnerStyle

	si := textinput.New()
	si.Placeholder = "search..."
	si.CharLimit = 100

	ai := textinput.New()
	ai.Placeholder = "reminder title"
	ai.CharLimit = 200

	ei := textinput.New()
	ei.CharLimit = 200

	return Model{
		client:      client,
		view:        ViewToday,
		mode:        ModeLoading,
		loading:     true,
		spinner:     s,
		searchInput: si,
		addInput:    ai,
		editInput:   ei,
		width:       80,
		height:      24,
	}
}

// Init starts the initial data load.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadData())
}

func (m Model) loadData() tea.Cmd {
	return func() tea.Msg {
		reminders, err := m.client.ShowAll()
		if err != nil {
			return remindersLoadedMsg{err: err}
		}
		lists, err := m.client.ListLists()
		if err != nil {
			return remindersLoadedMsg{reminders: reminders, err: err}
		}
		return remindersLoadedMsg{reminders: reminders, lists: lists}
	}
}

func statusTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return statusTickMsg{}
	})
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case remindersLoadedMsg:
		m.loading = false
		m.mode = ModeNormal
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
			m.statusErr = true
			m.statusTimer = 5
			// Still use whatever data we got
			if msg.reminders != nil {
				m.reminders = msg.reminders
			}
			m.buildDisplayItems()
			return m, statusTickCmd()
		}
		m.reminders = msg.reminders
		m.lists = msg.lists
		m.buildDisplayItems()
		return m, nil

	case actionDoneMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
			m.statusErr = true
			m.statusTimer = 5
			return m, statusTickCmd()
		}
		m.statusMsg = msg.msg
		m.statusErr = false
		m.statusTimer = 3
		// Refresh data after action
		m.loading = true
		return m, tea.Batch(statusTickCmd(), m.loadData(), m.spinner.Tick)

	case statusTickMsg:
		if m.statusTimer > 0 {
			m.statusTimer--
			if m.statusTimer == 0 {
				m.statusMsg = ""
				m.statusErr = false
			}
			return m, statusTickCmd()
		}
		return m, nil
	}

	// Mode-specific updates
	switch m.mode {
	case ModeSearch:
		return m.updateSearch(msg)
	case ModeAdd:
		return m.updateAdd(msg)
	case ModeAddField:
		return m.updateAddField(msg)
	case ModeEdit:
		return m.updateEdit(msg)
	case ModeEditField:
		return m.updateEditField(msg)
	case ModeConfirmDelete:
		return m.updateConfirmDelete(msg)
	case ModeHelp:
		return m.updateHelp(msg)
	default:
		return m.updateNormal(msg)
	}
}

func (m Model) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Up):
			m.moveCursor(-1)
		case key.Matches(msg, keys.Down):
			m.moveCursor(1)
		case key.Matches(msg, keys.Complete):
			return m.completeSelected()
		case key.Matches(msg, keys.Add):
			return m.enterAddMode()
		case key.Matches(msg, keys.Edit):
			return m.enterEditMode()
		case key.Matches(msg, keys.Delete):
			return m.enterDeleteConfirm()
		case key.Matches(msg, keys.Search):
			return m.enterSearchMode()
		case key.Matches(msg, keys.Refresh):
			m.loading = true
			m.mode = ModeLoading
			return m, tea.Batch(m.loadData(), m.spinner.Tick)
		case key.Matches(msg, keys.Help):
			m.mode = ModeHelp
		case key.Matches(msg, keys.Tab):
			m.view = (m.view + 1) % 4
			m.cursor = 0
			m.buildDisplayItems()
		case key.Matches(msg, keys.View1):
			m.view = ViewToday
			m.cursor = 0
			m.buildDisplayItems()
		case key.Matches(msg, keys.View2):
			m.view = ViewUpcoming
			m.cursor = 0
			m.buildDisplayItems()
		case key.Matches(msg, keys.View3):
			m.view = ViewLists
			m.cursor = 0
			m.buildDisplayItems()
		case key.Matches(msg, keys.View4):
			m.view = ViewAll
			m.cursor = 0
			m.buildDisplayItems()
		}
	}
	return m, nil
}

func (m *Model) moveCursor(delta int) {
	if len(m.displayItems) == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.displayItems) {
		m.cursor = len(m.displayItems) - 1
	}
	// Skip headers
	if m.displayItems[m.cursor].isHeader {
		m.cursor += delta
		if m.cursor < 0 {
			m.cursor = 0
		}
		if m.cursor >= len(m.displayItems) {
			m.cursor = len(m.displayItems) - 1
		}
	}
}

func (m Model) selectedReminder() (remindctl.Reminder, bool) {
	if m.cursor < 0 || m.cursor >= len(m.displayItems) {
		return remindctl.Reminder{}, false
	}
	item := m.displayItems[m.cursor]
	if item.isHeader {
		return remindctl.Reminder{}, false
	}
	return item.reminder, true
}

func (m Model) completeSelected() (tea.Model, tea.Cmd) {
	r, ok := m.selectedReminder()
	if !ok {
		return m, nil
	}
	client := m.client
	idPrefix := r.ID[:8]
	title := r.Title
	return m, func() tea.Msg {
		err := client.Complete(idPrefix)
		if err != nil {
			return actionDoneMsg{err: err}
		}
		return actionDoneMsg{msg: fmt.Sprintf("Completed: %s", ui.Truncate(title, 40))}
	}
}

func (m Model) enterAddMode() (tea.Model, tea.Cmd) {
	m.mode = ModeAdd
	m.addTitle = ""
	m.addList = ""
	m.addDue = ""
	m.addPriority = ""
	m.addStep = 0
	m.addInput.SetValue("")
	m.addInput.Placeholder = "reminder title"
	m.addInput.Focus()
	return m, m.addInput.Cursor.BlinkCmd()
}

func (m Model) updateAdd(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Escape):
			m.mode = ModeNormal
			m.addInput.Blur()
			return m, nil
		case key.Matches(msg, keys.Enter):
			title := strings.TrimSpace(m.addInput.Value())
			if title == "" {
				return m, nil
			}
			m.addTitle = title
			m.addStep = 1
			m.mode = ModeAddField
			m.addInput.SetValue("")
			m.addInput.Placeholder = "list name (enter to skip)"
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.addInput, cmd = m.addInput.Update(msg)
	return m, cmd
}

func (m Model) updateAddField(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Escape):
			m.mode = ModeNormal
			m.addInput.Blur()
			return m, nil
		case key.Matches(msg, keys.Enter):
			val := strings.TrimSpace(m.addInput.Value())
			switch m.addStep {
			case 1: // list
				m.addList = val
				m.addStep = 2
				m.addInput.SetValue("")
				m.addInput.Placeholder = "due date (enter to skip)"
				return m, nil
			case 2: // due
				m.addDue = val
				m.addStep = 3
				m.addInput.SetValue("")
				m.addInput.Placeholder = "priority: none/low/medium/high (enter to skip)"
				return m, nil
			case 3: // priority
				m.addPriority = val
				m.addInput.Blur()
				m.mode = ModeNormal
				return m, m.doAdd()
			}
		}
	}
	var cmd tea.Cmd
	m.addInput, cmd = m.addInput.Update(msg)
	return m, cmd
}

func (m Model) doAdd() tea.Cmd {
	client := m.client
	title := m.addTitle
	list := m.addList
	due := m.addDue
	priority := m.addPriority
	return func() tea.Msg {
		err := client.Add(title, list, due, priority)
		if err != nil {
			return actionDoneMsg{err: err}
		}
		return actionDoneMsg{msg: fmt.Sprintf("Added: %s", ui.Truncate(title, 40))}
	}
}

func (m Model) enterEditMode() (tea.Model, tea.Cmd) {
	r, ok := m.selectedReminder()
	if !ok {
		return m, nil
	}
	m.mode = ModeEdit
	m.editID = r.ID[:8]
	m.editStep = 0
	m.editTitle = ""
	m.editDue = ""
	m.editPriority = ""
	m.editInput.SetValue(r.Title)
	m.editInput.Placeholder = "new title (enter to keep)"
	m.editInput.Focus()
	return m, m.editInput.Cursor.BlinkCmd()
}

func (m Model) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Escape):
			m.mode = ModeNormal
			m.editInput.Blur()
			return m, nil
		case key.Matches(msg, keys.Enter):
			m.editTitle = strings.TrimSpace(m.editInput.Value())
			m.editStep = 1
			m.mode = ModeEditField
			m.editInput.SetValue("")
			m.editInput.Placeholder = "due date (enter to skip)"
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m Model) updateEditField(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Escape):
			m.mode = ModeNormal
			m.editInput.Blur()
			return m, nil
		case key.Matches(msg, keys.Enter):
			val := strings.TrimSpace(m.editInput.Value())
			switch m.editStep {
			case 1: // due
				m.editDue = val
				m.editStep = 2
				m.editInput.SetValue("")
				m.editInput.Placeholder = "priority: none/low/medium/high (enter to skip)"
				return m, nil
			case 2: // priority
				m.editPriority = val
				m.editInput.Blur()
				m.mode = ModeNormal
				return m, m.doEdit()
			}
		}
	}
	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m Model) doEdit() tea.Cmd {
	client := m.client
	id := m.editID
	title := m.editTitle
	due := m.editDue
	priority := m.editPriority
	return func() tea.Msg {
		err := client.Edit(id, title, due, priority)
		if err != nil {
			return actionDoneMsg{err: err}
		}
		return actionDoneMsg{msg: "Edited reminder"}
	}
}

func (m Model) enterDeleteConfirm() (tea.Model, tea.Cmd) {
	r, ok := m.selectedReminder()
	if !ok {
		return m, nil
	}
	m.mode = ModeConfirmDelete
	m.deleteID = r.ID[:8]
	m.deleteTitle = r.Title
	return m, nil
}

func (m Model) updateConfirmDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.mode = ModeNormal
			client := m.client
			id := m.deleteID
			title := m.deleteTitle
			return m, func() tea.Msg {
				err := client.Delete(id)
				if err != nil {
					return actionDoneMsg{err: err}
				}
				return actionDoneMsg{msg: fmt.Sprintf("Deleted: %s", ui.Truncate(title, 40))}
			}
		default:
			m.mode = ModeNormal
		}
	}
	return m, nil
}

func (m Model) enterSearchMode() (tea.Model, tea.Cmd) {
	m.mode = ModeSearch
	m.searchInput.SetValue(m.searchQuery)
	m.searchInput.Focus()
	return m, m.searchInput.Cursor.BlinkCmd()
}

func (m Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Escape):
			m.mode = ModeNormal
			m.searchQuery = ""
			m.searchInput.Blur()
			m.buildDisplayItems()
			return m, nil
		case key.Matches(msg, keys.Enter):
			m.mode = ModeNormal
			m.searchQuery = m.searchInput.Value()
			m.searchInput.Blur()
			m.cursor = 0
			m.buildDisplayItems()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	// Live filter as user types
	m.searchQuery = m.searchInput.Value()
	m.cursor = 0
	m.buildDisplayItems()
	return m, cmd
}

func (m Model) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Any key dismisses help
		_ = msg
		m.mode = ModeNormal
	}
	return m, nil
}

// buildDisplayItems computes the displayItems for the current view.
func (m *Model) buildDisplayItems() {
	now := time.Now()
	incomplete := m.incompleteReminders()

	// Apply search filter
	if m.searchQuery != "" {
		q := strings.ToLower(m.searchQuery)
		var filtered []remindctl.Reminder
		for _, r := range incomplete {
			if strings.Contains(strings.ToLower(r.Title), q) ||
				strings.Contains(strings.ToLower(r.ListName), q) ||
				strings.Contains(strings.ToLower(r.Notes), q) {
				filtered = append(filtered, r)
			}
		}
		incomplete = filtered
	}

	var items []displayItem

	switch m.view {
	case ViewToday:
		items = m.buildTodayItems(incomplete, now)
	case ViewUpcoming:
		items = m.buildUpcomingItems(incomplete, now)
	case ViewLists:
		items = m.buildListsItems(incomplete)
	case ViewAll:
		items = m.buildAllItems(incomplete, now)
	}

	m.displayItems = items

	// Fix cursor bounds
	if m.cursor >= len(m.displayItems) {
		m.cursor = len(m.displayItems) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	// Skip header if cursor landed on one
	if len(m.displayItems) > 0 && m.displayItems[m.cursor].isHeader {
		m.moveCursor(1)
	}
}

func (m *Model) incompleteReminders() []remindctl.Reminder {
	var result []remindctl.Reminder
	for _, r := range m.reminders {
		if !r.IsCompleted {
			result = append(result, r)
		}
	}
	return result
}

func (m *Model) buildTodayItems(reminders []remindctl.Reminder, now time.Time) []displayItem {
	var overdue, today, noDue []remindctl.Reminder
	for _, r := range reminders {
		if r.IsOverdue(now) {
			overdue = append(overdue, r)
		} else if r.IsDueToday(now) {
			today = append(today, r)
		} else if !r.HasDueDate() {
			noDue = append(noDue, r)
		}
	}

	sortByPriority(overdue)
	sortByPriority(today)

	var items []displayItem
	if len(overdue) > 0 {
		items = append(items, displayItem{isHeader: true, header: fmt.Sprintf("  Overdue (%d)", len(overdue)), overdue: true})
		for _, r := range overdue {
			items = append(items, displayItem{reminder: r})
		}
	}
	if len(today) > 0 {
		items = append(items, displayItem{isHeader: true, header: fmt.Sprintf("  Due Today (%d)", len(today))})
		for _, r := range today {
			items = append(items, displayItem{reminder: r})
		}
	}
	if len(noDue) > 0 {
		items = append(items, displayItem{isHeader: true, header: fmt.Sprintf("  No Due Date (%d)", len(noDue))})
		for _, r := range noDue {
			items = append(items, displayItem{reminder: r})
		}
	}
	return items
}

func (m *Model) buildUpcomingItems(reminders []remindctl.Reminder, now time.Time) []displayItem {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := today.AddDate(0, 0, 8) // 7 days from tomorrow

	// Group by date
	type dateGroup struct {
		date      time.Time
		reminders []remindctl.Reminder
	}
	groups := make(map[string]*dateGroup)
	var overdue []remindctl.Reminder

	for _, r := range reminders {
		if !r.HasDueDate() {
			continue
		}
		due := r.ParsedDueDate()
		dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, now.Location())

		if dueDay.Before(today) {
			overdue = append(overdue, r)
			continue
		}
		if dueDay.After(end) || dueDay.Equal(end) {
			continue
		}

		key := dueDay.Format("2006-01-02")
		if _, ok := groups[key]; !ok {
			groups[key] = &dateGroup{date: dueDay}
		}
		groups[key].reminders = append(groups[key].reminders, r)
	}

	// Sort groups by date
	var sortedGroups []*dateGroup
	for _, g := range groups {
		sortedGroups = append(sortedGroups, g)
	}
	sort.Slice(sortedGroups, func(i, j int) bool {
		return sortedGroups[i].date.Before(sortedGroups[j].date)
	})

	var items []displayItem

	if len(overdue) > 0 {
		sortByPriority(overdue)
		items = append(items, displayItem{isHeader: true, header: fmt.Sprintf("  Overdue (%d)", len(overdue)), overdue: true})
		for _, r := range overdue {
			items = append(items, displayItem{reminder: r})
		}
	}

	for _, g := range sortedGroups {
		sortByPriority(g.reminders)
		label := ui.RelativeDate(g.date, now)
		dateStr := g.date.Format("Mon Jan 2")
		if label == "today" || label == "tomorrow" {
			label = strings.ToUpper(label[:1]) + label[1:]
		} else {
			label = dateStr
		}
		items = append(items, displayItem{
			isHeader: true,
			header:   fmt.Sprintf("  %s (%d)", label, len(g.reminders)),
		})
		for _, r := range g.reminders {
			items = append(items, displayItem{reminder: r})
		}
	}

	return items
}

func (m *Model) buildListsItems(reminders []remindctl.Reminder) []displayItem {
	// Group by list name
	type listGroup struct {
		name      string
		reminders []remindctl.Reminder
	}
	groupMap := make(map[string]*listGroup)
	var order []string

	for _, r := range reminders {
		name := r.ListName
		if _, ok := groupMap[name]; !ok {
			groupMap[name] = &listGroup{name: name}
			order = append(order, name)
		}
		groupMap[name].reminders = append(groupMap[name].reminders, r)
	}

	sort.Strings(order)

	var items []displayItem
	for _, name := range order {
		g := groupMap[name]
		sortByPriority(g.reminders)
		items = append(items, displayItem{
			isHeader: true,
			header:   fmt.Sprintf("  %s (%d)", name, len(g.reminders)),
		})
		for _, r := range g.reminders {
			items = append(items, displayItem{reminder: r})
		}
	}
	return items
}

func (m *Model) buildAllItems(reminders []remindctl.Reminder, now time.Time) []displayItem {
	// Sort: overdue first, then by due date, then no-date
	sorted := make([]remindctl.Reminder, len(reminders))
	copy(sorted, reminders)

	sort.Slice(sorted, func(i, j int) bool {
		iOverdue := sorted[i].IsOverdue(now)
		jOverdue := sorted[j].IsOverdue(now)
		if iOverdue != jOverdue {
			return iOverdue
		}

		iDue := sorted[i].ParsedDueDate()
		jDue := sorted[j].ParsedDueDate()
		iHas := sorted[i].HasDueDate()
		jHas := sorted[j].HasDueDate()

		if iHas && jHas {
			return iDue.Before(jDue)
		}
		if iHas != jHas {
			return iHas
		}
		return sorted[i].Title < sorted[j].Title
	})

	var items []displayItem
	for _, r := range sorted {
		items = append(items, displayItem{reminder: r})
	}
	return items
}

func sortByPriority(reminders []remindctl.Reminder) {
	order := map[string]int{"high": 0, "medium": 1, "low": 2, "none": 3, "": 3}
	sort.SliceStable(reminders, func(i, j int) bool {
		return order[reminders[i].Priority] < order[reminders[j].Priority]
	})
}

// View renders the UI.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Tab bar
	b.WriteString(m.renderTabBar())
	b.WriteString("\n\n")

	// Main content
	if m.loading && len(m.reminders) == 0 {
		b.WriteString("  " + m.spinner.View() + " Loading reminders...")
	} else if m.mode == ModeHelp {
		b.WriteString(m.renderHelp())
	} else {
		b.WriteString(m.renderList())
	}

	// Input bar (for search/add/edit)
	if m.mode == ModeSearch {
		b.WriteString("\n")
		b.WriteString(m.renderSearchBar())
	} else if m.mode == ModeAdd || m.mode == ModeAddField {
		b.WriteString("\n")
		b.WriteString(m.renderAddBar())
	} else if m.mode == ModeEdit || m.mode == ModeEditField {
		b.WriteString("\n")
		b.WriteString(m.renderEditBar())
	} else if m.mode == ModeConfirmDelete {
		b.WriteString("\n")
		b.WriteString(m.renderDeleteConfirm())
	}

	// Pad to fill height
	lines := strings.Count(b.String(), "\n")
	for i := lines; i < m.height-2; i++ {
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m Model) renderHeader() string {
	title := ui.ViewTitleStyle.Render("  " + m.view.String())
	count := m.reminderCount()
	countStr := ui.CountStyle.Render(fmt.Sprintf(" (%d)", count))

	search := ""
	if m.searchQuery != "" && m.mode != ModeSearch {
		search = ui.CountStyle.Render(fmt.Sprintf("  filter: %q", m.searchQuery))
	}

	return title + countStr + search
}

func (m Model) reminderCount() int {
	count := 0
	for _, item := range m.displayItems {
		if !item.isHeader {
			count++
		}
	}
	return count
}

func (m Model) renderTabBar() string {
	tabs := []struct {
		view View
		key  string
		name string
	}{
		{ViewToday, "1", "Today"},
		{ViewUpcoming, "2", "Upcoming"},
		{ViewLists, "3", "Lists"},
		{ViewAll, "4", "All"},
	}

	var parts []string
	for _, tab := range tabs {
		label := fmt.Sprintf(" %s %s ", tab.key, tab.name)
		if tab.view == m.view {
			parts = append(parts, ui.ActiveTabStyle.Render(label))
		} else {
			parts = append(parts, ui.InactiveTabStyle.Render(label))
		}
	}

	sep := ui.TabSepStyle.Render(" ")
	return "  " + strings.Join(parts, sep)
}

func (m Model) renderList() string {
	if len(m.displayItems) == 0 {
		msg := "No reminders"
		switch m.view {
		case ViewToday:
			msg = "Nothing due today -- nice!"
		case ViewUpcoming:
			msg = "Nothing coming up this week"
		}
		if m.searchQuery != "" {
			msg = fmt.Sprintf("No results for %q", m.searchQuery)
		}
		return "  " + ui.CountStyle.Render(msg)
	}

	// Compute visible window
	maxVisible := m.height - 8 // header + tabs + status + padding
	if maxVisible < 5 {
		maxVisible = 5
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.displayItems) {
		end = len(m.displayItems)
	}

	now := time.Now()
	var b strings.Builder

	for i := start; i < end; i++ {
		item := m.displayItems[i]
		if item.isHeader {
			style := ui.GroupHeaderStyle
			if item.overdue {
				style = ui.GroupHeaderOverdueStyle
			}
			b.WriteString(style.Render(item.header))
			b.WriteString("\n")
			continue
		}

		line := m.renderReminder(item.reminder, now, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderReminder(r remindctl.Reminder, now time.Time, selected bool) string {
	maxTitle := m.width - 40
	if maxTitle < 20 {
		maxTitle = 20
	}

	// Priority dot
	dot := ui.PriorityIndicator(r.Priority)

	// Title
	title := ui.Truncate(r.Title, maxTitle)
	if r.IsOverdue(now) {
		title = ui.TitleOverdueStyle.Render(title)
	} else {
		title = ui.TitleStyle.Render(title)
	}

	// Due date
	dueStr := ""
	if r.HasDueDate() {
		due := r.ParsedDueDate()
		dueStr = " " + ui.FormatDueDate(due, now)
	}

	// List tag
	tag := ui.ListTagStyle(r.ListName).Render(" " + r.ListName)

	// Overdue badge
	badge := ""
	if r.IsOverdue(now) {
		badge = " " + ui.OverdueBadge()
	}

	line := fmt.Sprintf("  %s %s%s%s%s", dot, title, dueStr, tag, badge)

	if selected {
		// Apply selection highlight to the full line
		cursor := lipgloss.NewStyle().Foreground(ui.ColorAccent).Bold(true).Render("▸")
		line = fmt.Sprintf(" %s%s %s%s%s%s", cursor, dot, title, dueStr, tag, badge)
		line = ui.SelectedStyle.Render(line)
	}

	return line
}

func (m Model) renderSearchBar() string {
	label := ui.InputLabelStyle.Render("  / ")
	return label + m.searchInput.View()
}

func (m Model) renderAddBar() string {
	var label string
	switch m.addStep {
	case 0:
		label = "  Add title: "
	case 1:
		label = "  List: "
	case 2:
		label = "  Due: "
	case 3:
		label = "  Priority: "
	}
	return ui.InputLabelStyle.Render(label) + m.addInput.View()
}

func (m Model) renderEditBar() string {
	var label string
	switch m.editStep {
	case 0:
		label = "  Edit title: "
	case 1:
		label = "  Due: "
	case 2:
		label = "  Priority: "
	}
	return ui.InputLabelStyle.Render(label) + m.editInput.View()
}

func (m Model) renderDeleteConfirm() string {
	return ui.ConfirmStyle.Render(fmt.Sprintf(
		"  Delete %q? (y/n)",
		ui.Truncate(m.deleteTitle, 40),
	))
}

func (m Model) renderHelp() string {
	bindings := []struct {
		key  string
		desc string
	}{
		{"j/k", "navigate up/down"},
		{"x/space", "complete reminder"},
		{"a", "add new reminder"},
		{"e", "edit selected reminder"},
		{"d", "delete with confirmation"},
		{"1/2/3/4", "switch view (Today/Upcoming/Lists/All)"},
		{"Tab", "cycle views"},
		{"/", "search/filter"},
		{"r", "refresh data"},
		{"?", "toggle this help"},
		{"q", "quit"},
		{"Esc", "cancel current action"},
	}

	var b strings.Builder
	b.WriteString("  ")
	b.WriteString(ui.HelpTitleStyle.Render("Keybindings"))
	b.WriteString("\n\n")

	for _, bind := range bindings {
		k := ui.HelpKeyStyle.Render("  " + bind.key)
		d := ui.HelpDescStyle.Render(bind.desc)
		b.WriteString(k + d + "\n")
	}

	return b.String()
}

func (m Model) renderStatusBar() string {
	if m.statusMsg != "" {
		if m.statusErr {
			return ui.StatusErrorStyle.Render("  " + m.statusMsg)
		}
		return ui.StatusSuccessStyle.Render("  ✓ " + m.statusMsg)
	}

	// Show keybind hints
	hints := "  j/k navigate  x complete  a add  e edit  d delete  / search  ? help  q quit"
	if m.loading {
		hints = "  " + m.spinner.View() + " refreshing..."
	}
	return ui.StatusBarStyle.Render(hints)
}
