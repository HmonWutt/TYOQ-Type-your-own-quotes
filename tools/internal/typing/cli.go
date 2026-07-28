package typing

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	_ "modernc.org/sqlite"
)

var (
	correctStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#89DCEB"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	nextStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4")).
			Background(lipgloss.Color("#45475A")).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#585B70"))
	blueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Bold(true)
	magentaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7"))
	yellowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF"))
	redStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	headerStyle  = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#CBA6F7")).
			Padding(1, 3).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#F5C2E7")).Bold(true)
	badgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(lipgloss.Color("#F9E2AF")).
			Padding(0, 1)
	badgeLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086"))
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCCCEC")).Italic(true)
	resultHeaderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#89B4FA")).
				Padding(1, 3).
				Align(lipgloss.Center).
				Foreground(lipgloss.Color("#89DCEB")).Bold(true)
	resultBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#45475A")).
			Padding(1, 3)
)

const LIMIT = 500

type model struct {
	targetText     string
	typed          int
	typedChars     []rune
	errorIndices   map[int]bool
	startTime      time.Time
	endTime        time.Time
	done           bool
	waitingRestart bool
	quotes         []string
	customText     string
	isCustom       bool
	width          int
	height         int
	scrollOffset   int
	index          int
}

func loadQuotes() []string {
	db, err := sql.Open("sqlite", "../data/seed.db")
	if err != nil {
		return nil
	}
	defer db.Close()
	queryByLength := fmt.Sprintf("select text from quotes where word_count < %d limit %d", 100, LIMIT)
	rows, err := db.Query(queryByLength)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var quotes []string
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			continue
		}
		quotes = append(quotes, text)
	}
	return quotes
}

type inputPhase int

const (
	phaseTyping inputPhase = iota
	phaseRedirecting
)

type customInputModel struct {
	input     []rune
	width     int
	height    int
	cancelled bool
}

func newCustomInputModel() customInputModel {
	return customInputModel{width: 80, height: 24}
}

func (m customInputModel) Init() tea.Cmd {
	return nil
}

func (m customInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.PasteMsg:
		m.input = append(m.input, []rune(msg.Content)...)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		case "shift+enter":
			m.input = []rune{}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if msg.Text != "" {
				m.input = append(m.input, []rune(msg.Text)...)
			}
		}
	}
	return m, nil
}

func (m customInputModel) View() tea.View {
	w := max(m.width, 1)
	textWidth := min(w-4, 60)

	instruction := "Welcome to TYOQ. Paste your text below"
	promptBox := headerStyle.Width(textWidth).Render(instruction)

	preview := string(m.input)
	if preview == "" {
		preview = dimStyle.Render("(waiting for input...)")
	} else {
		preview = correctStyle.Width(textWidth).Render(preview)
	}

	footer := footerStyle.Render("enter to confirm · esc to quit . shift+enter to reset the text")

	body := lipgloss.JoinVertical(lipgloss.Center,
		promptBox,
		"",
		preview,
		"",
		footer,
	)
	content := lipgloss.PlaceHorizontal(w, lipgloss.Center, body)

	lines := strings.Count(content, "\n") + 1
	pad := max(m.height-lines, 0)

	v := tea.NewView(content + strings.Repeat("\n", pad))
	v.AltScreen = true
	return v
}

// runCustomInput launches the custom-input screen as its own bubbletea
// program and returns the text the user entered. Returns "" and exits
// the process if the user cancels with esc/ctrl+c.
func runCustomInput() string {
	p := tea.NewProgram(newCustomInputModel())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	m := finalModel.(customInputModel)
	if m.cancelled {
		os.Exit(0)
	}

	return strings.TrimSpace(string(m.input))
}

func initialModel() model {
	isCustom := false
	customText := ""
	var quotes []string

	if len(os.Args) > 1 && os.Args[1] == "-i" {
		isCustom = true
		customText = runCustomInput()
	} else {
		length, author := runQuoteSelection()
		quotes = loadQuotesFiltered(length, author)
		for len(quotes) <= 0 {
			length, author := runQuoteSelection()
			quotes = loadQuotesFiltered(length, author)
		}
	}

	targetText := ""
	if isCustom {
		targetText = customText
	} else if len(quotes) > 0 {
		targetText = quotes[0]
	}
	targetText = strings.Join(strings.Fields(targetText), " ")
	return model{
		targetText:   targetText,
		quotes:       quotes,
		customText:   customText,
		isCustom:     isCustom,
		errorIndices: make(map[int]bool),
		startTime:    time.Now(),
		width:        80,
		height:       24,
		index:        1,
	}
}

// loadQuotesFiltered loads quotes matching the given length bucket, and
// author. Pass "Any" (or "") for a filter to skip it.
func loadQuotesFiltered(length, author string) []string {
	db, err := sql.Open("sqlite", "../data/seed.db")
	if err != nil {
		return nil
	}
	defer db.Close()
	query := "select text from quotes"
	var conditions []string
	var args []any

	if author != "" && author != "Any" {
		conditions = append(conditions, "quotes.author like ?")
		args = append(args, author)
	}
	switch length {
	case "Short":
		conditions = append(conditions, "word_count <= 30")
	case "Medium":
		conditions = append(conditions, "word_count > 30 and word_count <= 50")
	case "Long":
		conditions = append(conditions, "word_count > 50 and word_count <=80")
	case "Extra Long":
		conditions = append(conditions, "word_count > 80 and word_count <=100")
	default:
		conditions = append(conditions, "")
	}

	if len(conditions) > 0 {
		query += " where " + strings.Join(conditions, " and ")
	}
	query += fmt.Sprintf(" limit %d", LIMIT)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var quotes []string
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			continue
		}
		quotes = append(quotes, text)
	}
	return quotes
}

type selectionStep int

const (
	stepLength selectionStep = iota
	stepAuthor
	stepDone
)

type quoteSelectionModel struct {
	step      selectionStep
	cursor    int
	scroll    int
	width     int
	height    int
	cancelled bool

	lengthOptions []string
	authorOptions []string

	chosenLength string
	chosenAuthor string
}

func newQuoteSelectionModel() quoteSelectionModel {
	authors := []string{"Any", "Douglas Adams", "Neil Gaiman", "Terry Pratchett"}
	return quoteSelectionModel{
		width:         80,
		height:        24,
		lengthOptions: []string{"Any", "Short", "Medium", "Long", "Extra Long"},
		authorOptions: authors,
	}
}

func (m quoteSelectionModel) currentOptions() []string {
	switch m.step {
	case stepLength:
		return m.lengthOptions
	case stepAuthor:
		return m.authorOptions
	}
	return nil
}

func (m quoteSelectionModel) stepTitle() string {
	switch m.step {
	case stepLength:
		return "Choose a length"
	case stepAuthor:
		return "Choose an author"
	}
	return ""
}

// advance moves to the next step, skipping any step that only has "Any".
func (m quoteSelectionModel) advance() (quoteSelectionModel, tea.Cmd) {
	m.step++
	m.cursor = 0
	m.scroll = 0
	if m.step < stepDone && len(m.currentOptions()) <= 1 {
		return m.advance()
	}
	if m.step >= stepDone {
		return m, tea.Quit
	}
	return m, nil
}

func (m quoteSelectionModel) Init() tea.Cmd {
	return nil
}

func (m quoteSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.currentOptions())-1 {
				m.cursor++
			}
		case "enter":
			choice := m.currentOptions()[m.cursor]
			switch m.step {
			case stepLength:
				m.chosenLength = choice
			case stepAuthor:
				m.chosenAuthor = choice
			}
			return m.advance()
		}
	}
	return m, nil
}

func (m quoteSelectionModel) View() tea.View {
	w := max(m.width, 1)
	boxWidth := min(w-4, 60)

	title := headerStyle.Width(boxWidth).Render(m.stepTitle())

	var breadcrumbs []string
	if m.chosenLength != "" {
		breadcrumbs = append(breadcrumbs, "Length: "+m.chosenLength)
	}
	if m.chosenAuthor != "" {
		breadcrumbs = append(breadcrumbs, "Author: "+m.chosenAuthor)
	}
	trail := dimStyle.Render(strings.Join(breadcrumbs, "  ·  "))

	const visible = 8
	options := m.currentOptions()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+visible {
		m.scroll = m.cursor - visible + 1
	}
	end := min(m.scroll+visible, len(options))

	var lines []string
	for i := m.scroll; i < end; i++ {
		label := options[i]
		if i == m.cursor {
			lines = append(lines, nextStyle.Render("> "+label))
		} else {
			lines = append(lines, dimStyle.Render("  "+label))
		}
	}
	list := strings.Join(lines, "\n")

	footer := footerStyle.Render("↑/↓ navigate · enter select · esc quit")

	body := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		trail,
		"",
		list,
		"",
		footer,
	)
	content := lipgloss.PlaceHorizontal(w, lipgloss.Center, body)

	lines2 := strings.Count(content, "\n") + 1
	pad := max(m.height-lines2, 0)

	v := tea.NewView(content + strings.Repeat("\n", pad))
	v.AltScreen = true
	return v
}

// runQuoteSelection launches the length/tag/author picker and returns the
// chosen filters. Returns ("", "", "") and exits if the user cancels.
func runQuoteSelection() (length, author string) {
	p := tea.NewProgram(newQuoteSelectionModel())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	m := finalModel.(quoteSelectionModel)
	if m.cancelled {
		os.Exit(0)
	}
	return m.chosenLength, m.chosenAuthor
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		if m.waitingRestart {
			switch msg.String() {
			case "esc", "ctrl+c":
				return m, tea.Quit
			default:
				return m.reset(m.width, m.height), nil
			}
		}
		if m.done {
			return m, nil
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m.reset(m.width, m.height), nil
		case "backspace":
			if m.typed > 0 {
				m.typed--
				m.typedChars = m.typedChars[:m.typed]
				delete(m.errorIndices, m.typed)
			}
		default:
			// Only handle actual character input (letters, space, punctuation, etc).
			// msg.Text is empty for non-character keys (enter, tab, arrows, ctrl combos,
			// function keys, etc), so this naturally filters those out instead of
			// accidentally typing their name (e.g. "enter") into the buffer.
			if msg.Text == "" {
				break
			}
			r := msg.Text
			targetRunes := []rune(m.targetText)
			if m.typed < len(targetRunes) {
				typedRunes := []rune(r)
				for _, tr := range typedRunes {
					if m.typed >= len(targetRunes) {
						break
					}
					expected := string(targetRunes[m.typed])
					if string(tr) != expected {
						m.errorIndices[m.typed] = true
					}
					m.typedChars = append(m.typedChars, tr)
					m.typed++
				}
				if m.typed >= len(targetRunes) {
					m.done = true
					m.waitingRestart = true
					m.endTime = time.Now()
				}
			}
		}
	}
	m.scrollOffset = m.updateScroll()
	return m, nil
}

func (m model) reset(width, height int) model {
	targetText := ""
	index := 1
	if m.isCustom {
		targetText = m.customText
	} else if len(m.quotes) > 0 {
		targetText = m.quotes[m.index]
		index = (m.index + 1) % len(m.quotes) // wrap around
	}
	targetText = strings.Join(strings.Fields(targetText), " ")
	return model{
		targetText:   targetText,
		quotes:       m.quotes,
		customText:   m.customText,
		isCustom:     m.isCustom,
		errorIndices: make(map[int]bool),
		startTime:    time.Now(),
		width:        width,
		height:       height,
		index:        index,
	}
}

type styleKind int

const (
	styleCorrect styleKind = iota
	styleError
	styleNext
	styleDim
)

func (m model) charStyleKind(i int) styleKind {
	if i < m.typed {
		if m.errorIndices[i] {
			return styleError
		}
		return styleCorrect
	}
	if i == m.typed && !m.done {
		return styleNext
	}
	return styleDim
}

func (m model) charStyle(kind styleKind) lipgloss.Style {
	switch kind {
	case styleError:
		return errorStyle
	case styleCorrect:
		return correctStyle
	case styleNext:
		return nextStyle
	default:
		return dimStyle
	}
}

func (m model) renderWrappedText(width, startLine, endLine int) string {
	runes := []rune(m.targetText)
	if width <= 0 {
		width = 80
	}
	var lines []string
	for lineNum := startLine; lineNum < endLine; lineNum++ {
		start := lineNum * width
		if start >= len(runes) {
			break
		}
		end := min(start+width, len(runes))
		var sb strings.Builder
		for i := start; i < end; i++ {
			ch := runes[i]
			kind := m.charStyleKind(i)
			style := m.charStyle(kind)
			if kind == styleError && ch == ' ' && i < len(m.typedChars) {
				sb.WriteString(style.Render(string(m.typedChars[i])))
			} else {
				sb.WriteString(style.Render(string(ch)))
			}
		}
		lines = append(lines, sb.String())
	}
	return strings.Join(lines, "\n")
}

func (m model) updateScroll() int {
	w := max(m.width, 1)
	cursorLine := m.typed / w
	totalLines := (len([]rune(m.targetText)) + w - 1) / w
	visible := m.textLinesVisible()
	cursorRow := 1
	offset := cursorLine - cursorRow
	if offset < 0 {
		offset = 0
	}
	maxOffset := totalLines - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	return offset
}

func (m model) textLinesVisible() int {
	totalLines := (len([]rune(m.targetText)) + max(m.width, 1) - 1) / max(m.width, 1)
	return min(3, totalLines)
}

func (m model) View() tea.View {
	var content string
	if m.waitingRestart {
		content = m.resultsView()
	} else {
		content = m.typingView()
	}
	lines := strings.Count(content, "\n") + 1
	pad := max(m.height-lines, 0)

	v := tea.NewView(content + strings.Repeat("\n", pad))
	v.AltScreen = true
	return v
}

func (m model) typingView() string {
	w := max(m.width, 1)
	header := "Let's see how fast you type!"
	headerBox := headerStyle.Width(min(w-4, 60)).Render(header)
	wordsTyped := 0
	if m.typed > 0 && m.typed <= len(m.targetText) {
		wordsTyped = strings.Count(m.targetText[:m.typed], " ")
	}
	totalWords := len(strings.Fields(m.targetText))
	badge := badgeLabelStyle.Render("words ") + badgeStyle.Render(fmt.Sprintf("%d/%d", wordsTyped, totalWords))
	startLine := m.scrollOffset
	endLine := startLine + m.textLinesVisible()
	textContent := m.renderWrappedText(w-2, startLine, endLine)
	footer := footerStyle.Render("esc to quit")
	body := lipgloss.JoinVertical(lipgloss.Center,
		headerBox,
		"",
		badge,
		"",
		textContent,
		"",
		footer,
	)
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, body)
}

func (m model) resultsView() string {
	w := max(m.width, 1)
	total := len([]rune(m.targetText))
	errors := len(m.errorIndices)

	var accuracy, speedChar, speedWord float64
	timeTaken := m.endTime.Sub(m.startTime).Seconds()
	if total > 0 && timeTaken > 0 {
		accuracy = float64(total-errors) / float64(total) * 100
		speedChar = float64(total) / timeTaken * 60
		speedWord = float64(total) / 5 / timeTaken * 60
	}

	header := "(⌐■_■) These are your results"
	headerBox := resultHeaderStyle.Width(min(w-4, 60)).Render(header)
	var stats strings.Builder
	rawLines := []string{
		fmt.Sprintf("Speed: %.0f wpm", speedWord),
		fmt.Sprintf("Speed: %.0f cpm", speedChar),
		fmt.Sprintf("Accuracy: %.0f%%", accuracy),
	}
	styles := []lipgloss.Style{blueStyle, magentaStyle, yellowStyle}
	contentW := 24
	for i, l := range rawLines {
		left := (contentW - len(l)) / 2
		right := contentW - len(l) - left
		padded := strings.Repeat(" ", left) + l + strings.Repeat(" ", right)
		stats.WriteString(styles[i].Render(padded))
		stats.WriteString("\n")
	}
	boxStyle := resultBoxStyle.Width(contentW + 8).Align(lipgloss.Center)
	statsBox := boxStyle.Render(stats.String())
	footer := footerStyle.Render("esc to quit · any other key to type again")
	body := lipgloss.JoinVertical(lipgloss.Center,
		headerBox,
		"",
		statsBox,
		"",
		footer,
	)
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, body)
}

func Type() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
