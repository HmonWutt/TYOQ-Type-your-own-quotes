package typing

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/db"
	_ "modernc.org/sqlite"
)

const tyoqBanner = `
$$$$$$$$\ $$\     $$\  $$$$$$\   $$$$$$\  
\__$$  __|\$$\   $$  |$$  __$$\ $$  __$$\ 
   $$ |    \$$\ $$  / $$ /  $$ |$$ /  $$ |
   $$ |     \$$$$  /  $$ |  $$ |$$ |  $$ |
   $$ |      \$$  /   $$ |  $$ |$$ |  $$ |
   $$ |       $$ |    $$ |  $$ |$$ $$\$$ |
   $$ |       $$ |     $$$$$$  |\$$$$$$ / 
   \__|       \__|     \______/  \___$$$\ 
                                     \___|`

var (
	correctStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#89DCEB"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	nextStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4")).
			Background(lipgloss.Color("#45475A")).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#585B70"))
	blueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Bold(true)
	magentaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	yellowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF"))
	redStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	greenStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	whiteStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	headerStyle  = lipgloss.NewStyle().
			Padding(1, 3).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#A6E3A1")).Bold(true)
	badgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(lipgloss.Color("#F9E2AF")).
			Padding(0, 1)
	badgeLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086"))
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCCCEC"))
	cardStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Align(lipgloss.Center)
	resultHeaderStyle = lipgloss.NewStyle().
				Padding(1, 3).
				Align(lipgloss.Center).
				Foreground(lipgloss.Color("#89DCEB")).Bold(true)
	resultBoxStyle = lipgloss.NewStyle().
			Padding(1, 3)
	statLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086"))
	hintKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(lipgloss.Color("#A6E3A1")).
			Padding(0, 1).Bold(true)
)

const LIMIT = 500

const (
	footerSep = "❖"
	statsSep  = "➮"
)

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
	menuRequested  bool
}

// sessionResult communicates what should happen after a typing program exits.
const (
	sessionQuit sessionResult = iota
	sessionMenu
)

type sessionResult int

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
	contentW := min(w-4, 64)

	instruction := "Welcome to TYOQ. Paste your text below"
	promptBox := headerStyle.Width(contentW).Render(instruction)

	preview := string(m.input)
	if preview == "" {
		preview = dimStyle.Render("(waiting for input...)")
	} else {
		preview = correctStyle.Render(preview)
	}
	previewCard := cardStyle.Width(contentW).Render(preview)

	footer := footerStyle.Padding(1, 1).Margin(1, 1).Render(highlightFooter("enter to confirm · esc to quit · shift+enter to reset the text"))

	body := lipgloss.JoinVertical(lipgloss.Center,
		promptBox,
		"",
		previewCard,
		"",
		footer,
	)
	content := lipgloss.PlaceHorizontal(w, lipgloss.Center, body)

	v := tea.NewView(content)
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

func initialModel(quotes []string, customText string, isCustom bool) model {
	targetText := ""
	index := 1
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
		index:        index,
	}
}

// loadQuotesFiltered loads quotes matching the given length bucket, and
// author. Pass "Any" (or "") for a filter to skip it.
func loadQuotesFiltered(length, author string) []string {
	path, err := db.EnsureDB()
	if err != nil {
		return nil
	}
	dbConn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil
	}
	defer dbConn.Close()
	query := "select text from quotes"
	var conditions []string
	var args []any

	if author != "" && author != "Any" {
		conditions = append(conditions, "quotes.author LIKE ?")
		author = fmt.Sprintf("%%%s%%", author)
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
		conditions = append(conditions, "word_count > 0")
	}

	if len(conditions) > 0 {
		query += " where " + strings.Join(conditions, " and ")
	}
	query += fmt.Sprintf(" limit %d", LIMIT)

	rows, err := dbConn.Query(query, args...)
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
		return "Choose length of quotes"
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
		return m, tea.ClearScreen

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "esc":
			if m.step == stepAuthor {
				m.step = stepLength
				m.cursor = 0
				m.scroll = 0
				return m, tea.ClearScreen
			}
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
			m, cmd := m.advance()
			return m, tea.Batch(cmd, tea.ClearScreen)
		}
	}
	return m, nil
}

func (m quoteSelectionModel) View() tea.View {
	w := max(m.width, 1)
	contentW := min(w-4, 64)

	title := headerStyle.Width(contentW).Render(m.stepTitle())

	var breadcrumbs []string
	if m.chosenLength != "" {
		breadcrumbs = append(breadcrumbs, "Length: "+m.chosenLength)
	}
	if m.chosenAuthor != "" {
		breadcrumbs = append(breadcrumbs, "Author: "+m.chosenAuthor)
	}
	trail := whiteStyle.Width(contentW).Align(lipgloss.Center).
		Render(strings.Join(breadcrumbs, "  "+footerSep+"  "))

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
			cursor := greenStyle.Bold(true).Render("❯")
			text := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A6E3A1")).
				Bold(true).
				Render(label)
			lines = append(lines, cursor+" "+text)
		} else {
			lines = append(lines, dimStyle.Render("  "+label))
		}
	}
	list := strings.Join(lines, "\n")
	listCard := cardStyle.Width(contentW).Render(list)

	footerText := "↑/↓ navigate · enter select · esc quit"
	if m.step == stepAuthor {
		footerText = "↑/↓ navigate · enter select · esc back · ctrl+c quit"
	}
	footer := footerStyle.Padding(1, 1).Margin(1, 1).Width(w).Align(lipgloss.Center).Render(highlightFooter(footerText))

	body := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		trail,
		"",
		listCard,
		"",
		footer,
	)
	content := lipgloss.PlaceHorizontal(w, lipgloss.Center, body)

	v := tea.NewView(content)
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
		return m, tea.ClearScreen
	case tea.KeyPressMsg:
		if m.waitingRestart {
			switch msg.String() {
			case "esc", "ctrl+c":
				return m, tea.Quit
			case "m":
				m.menuRequested = true
				return m, tea.Quit
			case "n":
				return m.reset(m.width, m.height), tea.ClearScreen
			case "r":
				return m.resetSame(m.width, m.height), tea.ClearScreen
			case "up", "down", "left", "right", "pgup", "pgdown", "home", "end":
				// ignore navigation keys (mouse wheel) on the results page
				return m, nil
			default:
				// any other key types the same quote again
				return m.resetSame(m.width, m.height), tea.ClearScreen
			}
		}
		if m.done {
			return m, nil
		}
		switch msg.String() {
		case "esc":
			m.menuRequested = true
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m.reset(m.width, m.height), tea.ClearScreen
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

// resetSame restarts the typing model with the same quote (retype).
func (m model) resetSame(width, height int) model {
	return model{
		targetText:   m.targetText,
		quotes:       m.quotes,
		customText:   m.customText,
		isCustom:     m.isCustom,
		errorIndices: make(map[int]bool),
		startTime:    time.Now(),
		width:        width,
		height:       height,
		index:        m.index,
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

// wrapLines splits targetText into visual lines of at most width runes,
// breaking at word boundaries when possible so words don't split mid-word.
// Each returned entry is the slice of rune indices belonging to that line.
func wrapLines(target string, width int) [][]int {
	runes := []rune(target)
	if width <= 0 {
		width = 80
	}
	if len(runes) == 0 {
		return [][]int{{}}
	}

	var lines [][]int
	var current []int

	flush := func() {
		lines = append(lines, current)
		current = nil
	}

	i := 0
	for i < len(runes) {
		// gather a word (run of non-space) or a single space
		isSpace := runes[i] == ' '
		j := i
		for j < len(runes) && (runes[j] == ' ') == isSpace {
			j++
		}
		token := make([]int, j-i)
		for k := i; k < j; k++ {
			token[k-i] = k
		}

		if len(current)+len(token) <= width {
			current = append(current, token...)
		} else {
			// word longer than a full line: hard-break it across lines
			if isSpace && len(current) == 0 {
				// leading space on empty line that doesn't fit: drop it
			} else if len(token) > width && !isSpace {
				remaining := token
				for len(remaining) > 0 {
					take := width - len(current)
					if take <= 0 {
						flush()
						continue
					}
					if take > len(remaining) {
						take = len(remaining)
					}
					current = append(current, remaining[:take]...)
					remaining = remaining[take:]
					if len(remaining) > 0 {
						flush()
					}
				}
			} else {
				flush()
				current = append(current, token...)
			}
		}
		i = j
	}
	flush()
	if len(lines) == 0 {
		lines = append(lines, nil)
	}
	return lines
}

// renderLines renders the given wrapped lines (each a slice of rune indices)
// applying per-character styling. Each line is padded to width so that
// shorter lines overwrite stale characters from the previous frame when
// scrolling.
func (m model) renderLines(lines [][]int, width int) string {
	if width <= 0 {
		width = 80
	}
	var out strings.Builder
	for li, line := range lines {
		if li > 0 {
			out.WriteString("\n")
		}
		used := 0
		for _, i := range line {
			ch := []rune(m.targetText)[i]
			kind := m.charStyleKind(i)
			style := m.charStyle(kind)
			if kind == styleError && ch == ' ' && i < len(m.typedChars) {
				out.WriteString(style.Render(string(m.typedChars[i])))
			} else {
				out.WriteString(style.Render(string(ch)))
			}
			used++
		}
		if used < width {
			out.WriteString(strings.Repeat(" ", width-used))
		}
	}
	return out.String()
}

// textWidth is the inner wrap width used for the typing text. It must
// match the width the text is actually rendered at in typingView so the
// scroll offset and visible-line count stay in sync with what is drawn.
func (m model) textWidth() int {
	w := max(m.width, 1)
	contentW := max(w*70/100, 1)
	return contentW
}

func (m model) updateScroll() int {
	tw := m.textWidth()
	lines := wrapLines(m.targetText, tw)
	totalLines := len(lines)
	visible := m.textLinesVisible()
	if visible <= 0 {
		return 0
	}
	// find which line the cursor (typed position) is on
	cursorLine := 0
	for li, line := range lines {
		for _, idx := range line {
			if idx == m.typed {
				cursorLine = li
				break
			}
		}
	}
	// keep cursor on the second visible row when possible
	cursorRow := min(1, visible-1)
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

// textLinesVisible returns how many wrapped text lines can be shown at
// once. It is capped at 3 but also reduced so the total typing-view
// height (header, badge, bar, text, footer) fits within the terminal
// height, which prevents the alt screen from scrolling and exposing
// previous frames.
func (m model) textLinesVisible() int {
	const maxVisible = 6
	// chrome lines: header (1), bar (1), footer (1),
	// plus blank separators between sections (5).
	const chrome = 8
	avail := m.height - chrome
	if avail < 1 {
		avail = 1
	}
	visible := min(maxVisible, avail)
	totalLines := len(wrapLines(m.targetText, m.textWidth()))
	return min(visible, totalLines)
}

func (m model) View() tea.View {
	var content string
	if m.waitingRestart {
		content = m.resultsView()
	} else {
		content = m.typingView()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m model) typingView() string {
	w := max(m.width, 1)
	contentW := max(w*70/100, 1)

	header := "Let's see how fast you type!"
	headerBox := headerStyle.Width(contentW).Render(header)

	// wordsTyped := 0
	// if m.typed > 0 && m.typed <= len(m.targetText) {
	// 	wordsTyped = strings.Count(m.targetText[:m.typed], " ")
	// }
	// totalWords := len(strings.Fields(m.targetText))
	// if totalWords <= 0 {
	// 	totalWords = 1
	// }
	// badge := badgeLabelStyle.Render("words ") +
	// 	badgeStyle.Render(fmt.Sprintf("%d/%d", wordsTyped, totalWords))
	bar := lipgloss.PlaceHorizontal(contentW, lipgloss.Center,
		progressBar(m.typed, len(m.targetText), min(contentW-4, 50)))

	allLines := wrapLines(m.targetText, m.textWidth())
	startLine := m.scrollOffset
	endLine := startLine + m.textLinesVisible()
	if startLine > len(allLines) {
		startLine = len(allLines)
	}
	if endLine > len(allLines) {
		endLine = len(allLines)
	}
	if endLine < startLine {
		endLine = startLine
	}
	textContent := m.renderLines(allLines[startLine:endLine], m.textWidth())

	footer := footerStyle.Padding(1, 1).Margin(1, 1).Width(w).Align(lipgloss.Center).
		Render(highlightFooter("enter next quote · esc back to menu · ctrl+c quit"))
	body := lipgloss.JoinVertical(lipgloss.Center,
		headerBox,
		bar,
		// badge,
		// "",
		"",
		textContent,
		"",
		"",
		footer,
	)
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, body)
}

func (m model) resultsView() string {
	w := max(m.width, 1)
	contentW := min(w-4, 64)
	total := len([]rune(m.targetText))
	errors := len(m.errorIndices)

	var accuracy, speedWord float64
	timeTaken := m.endTime.Sub(m.startTime).Seconds()
	if total > 0 && timeTaken > 0 {
		accuracy = float64(total-errors) / float64(total) * 100
		speedWord = float64(total) / 5 / timeTaken * 60
	}

	header := "(⌐■_■) These are your results"
	headerBox := resultHeaderStyle.Width(contentW).Render(header)

	// stat rows: label left, value right, inside a bordered card
	statRow := func(label, value string, vStyle lipgloss.Style) string {
		labelRender := whiteStyle.Render(label)
		valueRender := vStyle.Render(value)
		// gap := max(contentW-16-unisafewidth(label)-unisafewidth(value), 1)
		return labelRender + strings.Repeat(" ", 1) + valueRender
	}
	rows := strings.Join([]string{
		statRow("Speed", fmt.Sprintf("%s %.0f wpm\n", footerSep, speedWord), greenStyle),
		statRow("Accuracy", fmt.Sprintf("%s %.0f%%\n", footerSep, accuracy), greenStyle),
		statRow("Time", fmt.Sprintf("%s %.1fs", footerSep, timeTaken), greenStyle),
	}, "\n")
	statsBox := cardStyle.Width(contentW).Align(lipgloss.Center).Render(rows)

	footer := greenStyle.Padding(1, 1).Margin(1, 1).Width(w).Align(lipgloss.Center).
		Render(highlightFooter("r repeat · n next quote · m menu · esc quit"))
	body := lipgloss.JoinVertical(lipgloss.Center,
		headerBox,
		"",
		statsBox,
		"",
		footer,
	)
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, body)
}

// unisafewidth returns the display width of s ignoring ANSI escapes.
// lipgloss.Width does this too, but we need it for raw strings before
// they get styled. Falls back to rune count.
func unisafewidth(s string) int {
	return lipgloss.Width(s)
}

// highlightFooter renders a "key1 desc1 · key2 desc2 · ..." footer string
// with each key highlighted via hintKeyStyle and the rest in footerStyle.
func highlightFooter(s string) string {
	parts := strings.Split(s, " · ")
	var rendered []string
	for _, p := range parts {
		fields := strings.SplitN(p, " ", 2)
		if len(fields) == 2 {
			rendered = append(rendered,
				hintKeyStyle.Render(fields[0])+" "+
					footerStyle.Render(fields[1]))
		} else {
			rendered = append(rendered, footerStyle.Render(p))
		}
	}
	return strings.Join(rendered, " "+dimStyle.Render(footerSep)+" ")
}

// progressBar renders a labelled progress bar of the given width.
func progressBar(current, total, width int) string {
	if total <= 0 {
		total = 1
	}
	if width < 10 {
		width = 10
	}
	pct := float64(current) / float64(total)
	if pct < 0 {
		pct = 0
	} else if pct > 1 {
		pct = 1
	}
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	bar := greenStyle.Render(strings.Repeat("▁", filled)) +
		whiteStyle.Render(strings.Repeat("▁", width-filled))
	return bar
}

func Type() {
	isCustom := len(os.Args) > 1 && os.Args[1] == "-i"

	runWelcome()

	for {
		quotes, customText := launchSelection(isCustom)
		if len(quotes) == 0 && customText == "" {
			return
		}

		m := initialModel(quotes, customText, isCustom)
		p := tea.NewProgram(m)
		finalModel, err := p.Run()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		result := finalModel.(model)
		if result.menuRequested {
			continue
		}
		return
	}
}

// launchSelection runs the appropriate selection screen (custom input or
// quote picker) and returns the quotes/customText to type. Returns empty
// values if the user cancelled.
func launchSelection(isCustom bool) (quotes []string, customText string) {
	if isCustom {
		return nil, runCustomInput()
	}
	for {
		length, author := runQuoteSelection()
		quotes = loadQuotesFiltered(length, author)
		if len(quotes) > 0 {
			return quotes, ""
		}
	}
}

type welcomeModel struct {
	width    int
	height   int
	progress int // 0..100
	done     bool
}

type tickMsg time.Time

func newWelcomeModel() welcomeModel {
	return welcomeModel{width: 80, height: 24}
}

func (m welcomeModel) Init() tea.Cmd {
	return tea.Tick(30*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m welcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		m.progress += 5
		if m.progress >= 100 {
			m.progress = 100
			m.done = true
			return m, tea.Quit
		}
		return m, tea.Tick(30*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			os.Exit(0)
		default:
			m.progress = 100
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m welcomeModel) View() tea.View {
	w := max(m.width, 1)
	contentW := min(w-4, 64)

	banner := magentaStyle.Width(contentW).Align(lipgloss.Center).Render(tyoqBanner)
	bar := progressBar(m.progress, 100, min(contentW-4, 40))

	body := lipgloss.JoinVertical(lipgloss.Center,
		banner,
		"",
		bar,
	)
	content := lipgloss.PlaceHorizontal(w, lipgloss.Center, body)

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// runWelcome shows the welcome screen with a loading progress bar and
// blocks until the bar completes or the user presses a key
// (esc/ctrl+c exits the process).
func runWelcome() {
	p := tea.NewProgram(newWelcomeModel())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	_ = finalModel
}
