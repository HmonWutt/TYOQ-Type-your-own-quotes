package typing

import (
	"bufio"
	"database/sql"
	"fmt"
	"math/rand/v2"
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

	headerStyle = lipgloss.NewStyle().
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
			Foreground(lipgloss.Color("#585B70")).Italic(true)

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
}

func loadQuotes() []string {
	db, err := sql.Open("sqlite", "../data/seed.db")
	if err != nil {
		return nil
	}
	defer db.Close()
	query := "SELECT text FROM quotes WHERE word_count < 50 LIMIT 5"
	rows, err := db.Query(query)
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

func readCustomInput() string {
	instruction := "Welcome to TYOQ. Paste your text below"
	length := len(instruction) + 4
	hLine := "+" + strings.Repeat("-", length) + "+"
	vLine := "|" + strings.Repeat(" ", length) + "|"
	vLineMid := "|  " + instruction + "  |"

	fmt.Println("\033[92m" + hLine + "\033[00m")
	fmt.Println("\033[92m" + vLine + "\033[00m")
	fmt.Println("\033[92m" + vLineMid + "\033[00m")
	fmt.Println("\033[92m" + vLine + "\033[00m")
	fmt.Println("\033[92m" + hLine + "\033[00m")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	fmt.Print("\033[92mInput saved. Redirecting to the typing arena \033[00m")
	for range 4 {
		fmt.Print("\033[92m➤\033[00m")
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println()
	return input
}

func initialModel() model {
	isCustom := false
	customText := ""

	if len(os.Args) > 1 && os.Args[1] == "-i" {
		isCustom = true
		customText = readCustomInput()
	}

	quotes := loadQuotes()
	targetText := ""
	if isCustom {
		targetText = customText
	} else if len(quotes) > 0 {
		targetText = quotes[rand.IntN(len(quotes))]
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
	}
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
		case "backspace":
			if m.typed > 0 {
				m.typed--
				m.typedChars = m.typedChars[:m.typed]
				delete(m.errorIndices, m.typed)
			}
		default:
			r := msg.String()
			if r == "space" {
				r = " "
			}
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
	if m.isCustom {
		targetText = m.customText
	} else if len(m.quotes) > 0 {
		targetText = m.quotes[rand.IntN(len(m.quotes))]
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
	return tea.NewView(content + strings.Repeat("\n", pad))
}

func (m model) typingView() string {
	w := max(m.width, 1)

	header := "Let's see how fast you can type!"
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
	accuracy := float64(total-errors) / float64(total) * 100
	timeTaken := m.endTime.Sub(m.startTime).Seconds()

	speedChar := float64(total) / timeTaken * 60
	speedWord := float64(total) / 5 / timeTaken * 60

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
