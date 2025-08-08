package tui

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DockStats holds a single container stat line parsed from `docker stats --no-stream --format`.
type DockStats struct {
	Name     string
	CPU      float64
	MemUsage string
	MemPerc  float64
	NetIO    string
	BlockIO  string
}

// model for Bubble Tea
type monitorModel struct {
	loading    bool
	paused     bool
	stats      []DockStats
	history    map[string][]float64 // cpu history (0..100)
	netHist    map[string][]float64 // combined up+down bytes/sec
	diskHist   map[string][]float64 // combined r+w bytes/sec
	prevTotals map[string]prevTotals
	maxPoints  int // sparkline width
	width      int
	height     int
	err        error
	idx        int
	selected   int
	showHelp   bool
}

type prevTotals struct {
	netRx float64
	netTx float64
	blkR  float64
	blkW  float64
	ts    time.Time
}

type tickMsg time.Time
type statsMsg struct {
	rows []DockStats
	ts   time.Time
}
type errMsg struct{ err error }

func newModel() monitorModel {
	return monitorModel{
		loading:    true,
		paused:     false,
		stats:      nil,
		history:    make(map[string][]float64),
		netHist:    make(map[string][]float64),
		diskHist:   make(map[string][]float64),
		prevTotals: make(map[string]prevTotals),
		maxPoints:  32,
		selected:   0,
		showHelp:   false,
	}
}

func (m monitorModel) Init() tea.Cmd {
	return tea.Batch(fetchStatsCmd(), tick())
}

func tick() tea.Cmd { return tea.Tick(1*time.Second, func(t time.Time) tea.Msg { return tickMsg(t) }) }

func fetchStatsCmd() tea.Cmd {
	return func() tea.Msg {
		rows, err := fetchOnce()
		if err != nil {
			return errMsg{err}
		}
		return statsMsg{rows: rows, ts: time.Now()}
	}
}

func (m monitorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case " ":
			m.paused = !m.paused
			return m, nil
		case "r":
			return m, fetchStatsCmd()
		case "h":
			m.showHelp = !m.showHelp
			return m, nil
		case "tab", "shift+tab", "down", "j":
			if len(m.stats) > 0 {
				m.selected = (m.selected + 1) % len(m.stats)
			}
			return m, nil
		case "up", "k":
			if len(m.stats) > 0 {
				m.selected = (m.selected - 1 + len(m.stats)) % len(m.stats)
			}
			return m, nil
		}
	case tickMsg:
		if !m.paused {
			return m, fetchStatsCmd()
		}
		return m, tick()
	case statsMsg:
		m.loading = false
		m.stats = msg.rows
		// update history
		for _, s := range m.stats {
			h := m.history[s.Name]
			h = append(h, s.CPU)
			if len(h) > 30 { // keep last 30 points
				h = h[len(h)-30:]
			}
			m.history[s.Name] = h
			// NET/DISK instantaneous rates from cumulative totals
			rx, tx := parseTwoBytes(s.NetIO)
			br, bw := parseTwoBytes(s.BlockIO)
			prev := m.prevTotals[s.Name]
			if !prev.ts.IsZero() {
				dt := msg.ts.Sub(prev.ts).Seconds()
				if dt > 0 {
					upRate := maxFloat((tx-prev.netTx)/dt, 0)
					downRate := maxFloat((rx-prev.netRx)/dt, 0)
					nh := m.netHist[s.Name]
					nh = append(nh, upRate+downRate)
					if len(nh) > m.maxPoints {
						nh = nh[len(nh)-m.maxPoints:]
					}
					m.netHist[s.Name] = nh

					rRate := maxFloat((br-prev.blkR)/dt, 0)
					wRate := maxFloat((bw-prev.blkW)/dt, 0)
					dh := m.diskHist[s.Name]
					dh = append(dh, rRate+wRate)
					if len(dh) > m.maxPoints {
						dh = dh[len(dh)-m.maxPoints:]
					}
					m.diskHist[s.Name] = dh
				}
			}
			m.prevTotals[s.Name] = prevTotals{netRx: rx, netTx: tx, blkR: br, blkW: bw, ts: msg.ts}
		}
		return m, tick()
	case errMsg:
		m.err = msg.err
		return m, tick()
	}
	return m, nil
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	boxStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	selStyle   = lipgloss.NewStyle().Bold(true).Underline(true)
	nameStyle  = lipgloss.NewStyle().Bold(true)
	cpuColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	memColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	netColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("45"))
	diskColor  = lipgloss.NewStyle().Foreground(lipgloss.Color("178"))
	barFilled  = "█"
	barEmpty   = " "
	sparklCols = []rune("▁▂▃▄▅▆▇█")
)

func (m monitorModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v\n", m.err)
	}
	if m.loading {
		return "Loading docker stats... (q to quit)\n"
	}
	var b strings.Builder
	// Header with selection info and time
	header := titleStyle.Render(fmt.Sprintf("Container Monitor   [%d/%d]  %s",
		clamp(m.selected+1, 1, len(m.stats)), len(m.stats), time.Now().Format("15:04:05")))
	b.WriteString(header)
	b.WriteString("\n")
	{
		for i, s := range m.stats {
			if i > 0 {
				b.WriteString("\n")
			}
			name := "🐳   " + s.Name
			if i == m.selected {
				name = selStyle.Render(name)
			} else {
				name = nameStyle.Render(name)
			}
			header := fmt.Sprintf("%s", name)
			b.WriteString(header)
			b.WriteString("\n")

			// CPU bar
			b.WriteString(cpuColor.Render("CPU:"))
			b.WriteString("  ")
			b.WriteString(progressBar(s.CPU, 10))
			avg := average(m.history[s.Name])
			b.WriteString(fmt.Sprintf(" %2.0f%%  ", s.CPU))
			b.WriteString(sparkline(m.history[s.Name], m.maxPoints))
			if avg >= 0 {
				b.WriteString(fmt.Sprintf(" (avg: %2.0f%%)", avg))
			}
			b.WriteString("\n")

			// MEM bar (only percentage, plus raw usage as-is)
			b.WriteString(memColor.Render("MEM:"))
			b.WriteString("  ")
			b.WriteString(progressBar(s.MemPerc, 10))
			b.WriteString(fmt.Sprintf(" %2.0f%%  %s\n", s.MemPerc, s.MemUsage))

			// NET instantaneous rate (from prev totals) + sparkline
			rx, tx := parseTwoBytes(s.NetIO)
			prev := m.prevTotals[s.Name]
			var upRate, downRate float64
			if !prev.ts.IsZero() {
				dt := time.Since(prev.ts).Seconds()
				if dt > 0 {
					upRate = maxFloat((tx-prev.netTx)/dt, 0)
					downRate = maxFloat((rx-prev.netRx)/dt, 0)
				}
			}
			b.WriteString(netColor.Render("NET:"))
			b.WriteString("  ")
			b.WriteString(fmt.Sprintf("↑%s/s ↓%s/s  ", humanBytes(upRate), humanBytes(downRate)))
			b.WriteString(sparkline(m.netHist[s.Name], m.maxPoints))
			b.WriteString("\n")
			// DISK instantaneous rate + sparkline
			br, bw := parseTwoBytes(s.BlockIO)
			var rRate, wRate float64
			if !prev.ts.IsZero() {
				dt := time.Since(prev.ts).Seconds()
				if dt > 0 {
					rRate = maxFloat((br-prev.blkR)/dt, 0)
					wRate = maxFloat((bw-prev.blkW)/dt, 0)
				}
			}
			b.WriteString(diskColor.Render("DISK:"))
			b.WriteString(" ")
			b.WriteString(fmt.Sprintf("R:%s/s W:%s/s  ", humanBytes(rRate), humanBytes(wRate)))
			b.WriteString(sparkline(m.diskHist[s.Name], m.maxPoints))
			b.WriteString("\n")
		}
	}
	if m.showHelp {
		b.WriteString("\n")
		help := boxStyle.Render("[↑/↓|TAB] Switch   [SPACE] Pause   [R] Refresh   [H] Help   [Q] Quit\n" +
			"- CPU/MEM: バーは使用率、CPUは右に履歴スパークラインと平均\n" +
			"- NET/DISK: 瞬間転送量を算出し、右側にスパークライン表示")
		b.WriteString(help)
	} else {
		b.WriteString("\n[↑/↓|TAB] Switch  [SPACE] Pause  [R] Refresh  [H] Help  [Q] Quit\n")
	}
	return b.String()
}

func progressBar(perc float64, cells int) string {
	if perc < 0 {
		perc = 0
	}
	if perc > 100 {
		perc = 100
	}
	filled := int((perc/100.0)*float64(cells) + 0.5)
	if filled > cells {
		filled = cells
	}
	return "[" + strings.Repeat(barFilled, filled) + strings.Repeat(barEmpty, cells-filled) + "]"
}

func sparkline(vals []float64, width int) string {
	if width <= 0 {
		width = 16
	}
	if len(vals) == 0 {
		return strings.Repeat(string(sparklCols[0]), width)
	}
	// take last width points; left pad with zeros if not enough
	var seg []float64
	if len(vals) > width {
		seg = vals[len(vals)-width:]
	} else {
		seg = make([]float64, width)
		pad := width - len(vals)
		for i := 0; i < pad; i++ {
			seg[i] = 0
		}
		copy(seg[pad:], vals)
	}
	// relative scaling within the segment
	minV, maxV := seg[0], seg[0]
	for _, v := range seg {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	if maxV-minV < 0.001 {
		return strings.Repeat(string(sparklCols[0]), width)
	}
	var out strings.Builder
	for _, v := range seg {
		norm := (v - minV) / (maxV - minV)
		idx := int(norm*float64(len(sparklCols)-1) + 0.00001)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparklCols) {
			idx = len(sparklCols) - 1
		}
		out.WriteRune(sparklCols[idx])
	}
	return out.String()
}

func average(vals []float64) float64 {
	if len(vals) == 0 {
		return -1
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

// bigChart renders a larger ASCII chart using 8 vertical levels, width W
func bigChart(vals []float64, width, height int) string {
	if height <= 1 {
		height = 8
	}
	if width <= 0 {
		width = 48
	}
	// prepare segment similar to sparkline
	var seg []float64
	if len(vals) > width {
		seg = vals[len(vals)-width:]
	} else {
		seg = make([]float64, width)
		copy(seg[width-len(vals):], vals)
	}
	minV, maxV := math.MaxFloat64, -math.MaxFloat64
	for _, v := range seg {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	if !isFinite(minV) || !isFinite(maxV) || maxV-minV < 1e-9 {
		minV, maxV = 0, 1
	}
	// draw from top row to bottom
	var b strings.Builder
	for row := height - 1; row >= 0; row-- {
		for _, v := range seg {
			norm := (v - minV) / (maxV - minV)
			lvl := int(norm*float64(height-1) + 0.00001)
			if lvl >= row {
				b.WriteRune('⣿')
			} else {
				b.WriteRune(' ')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func isFinite(f float64) bool { return !(math.IsNaN(f) || math.IsInf(f, 0)) }

// chartWithAxis left for compatibility (not used in List mode now)
func chartWithAxis(vals []float64, width, height int, unit string) string { return "" }

// chartLineWithAxis renders a smooth-looking line using Unicode braille characters.
// Each cell encodes a 2x4 dot matrix to achieve sub-row resolution (similarにLazydockerの滑らかさを再現)。
func chartLineWithAxis(vals []float64, width, height int, unit string) string {
	if width < 20 {
		width = 20
	}
	if height < 6 {
		height = 6
	}

	// segment
	seg := vals
	if len(seg) > width {
		seg = seg[len(seg)-width:]
	} else {
		tmp := make([]float64, width)
		copy(tmp[width-len(seg):], seg)
		seg = tmp
	}
	// scale
	minV, maxV := math.MaxFloat64, -math.MaxFloat64
	for _, v := range seg {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	if !isFinite(minV) || !isFinite(maxV) || maxV-minV < 1e-9 {
		minV, maxV = 0, 1
	}
	// add padding to avoid jitter when values are flat
	rng := maxV - minV
	pad := rng * 0.1
	if pad < 1e-6 {
		pad = 1
	}
	minV -= pad
	maxV += pad

	// smoothing (EMA)
	alpha := 0.35
	last := seg[0]
	for i := 1; i < len(seg); i++ {
		last = alpha*seg[i] + (1-alpha)*last
		seg[i] = last
	}

	cols := width
	rows := height
	grid := make([][]rune, rows)
	for r := range grid {
		grid[r] = make([]rune, cols)
		for c := range grid[r] {
			grid[r][c] = ' '
		}
	}
	// guides
	for _, gy := range []int{0, rows / 2, rows - 1} {
		for c := 0; c < cols; c++ {
			grid[gy][c] = '┈'
		}
	}

	yOf := func(v float64) int {
		norm := (v - minV) / (maxV - minV)
		y := int((1-norm)*float64(rows-1) + 0.5)
		if y < 0 {
			y = 0
		}
		if y >= rows {
			y = rows - 1
		}
		return y
	}
	prevY := yOf(seg[0])
	for x := 1; x < cols; x++ {
		y := yOf(seg[x])
		x0, y0, x1, y1 := x-1, prevY, x, y
		dx := x1 - x0
		dy := y1 - y0
		sy := 1
		if dy < 0 {
			dy = -dy
			sy = -1
		}
		err := dx / 2
		cy := y0
		for cx := x0; cx <= x1; cx++ {
			putLineChar(grid, cx, cy, x0, y0, x1, y1)
			err -= dy
			if err < 0 {
				cy += sy
				err += dx
			}
		}
		prevY = y
	}

	tickW := 8
	var b strings.Builder
	for r := 0; r < rows; r++ {
		label := ""
		switch r {
		case 0:
			label = formatTick(maxV, unit)
		case rows / 2:
			label = formatTick((maxV+minV)/2, unit)
		case rows - 1:
			label = formatTick(minV, unit)
		}
		b.WriteString(padLeft(label, tickW))
		b.WriteString(" ")
		for c := 0; c < cols; c++ {
			b.WriteRune(grid[r][c])
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func putLineChar(grid [][]rune, x, y, x0, y0, x1, y1 int) {
	if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[0]) {
		return
	}
	ch := '─'
	if y != y0 {
		if (y1-y0)*(x1-x0) > 0 {
			ch = '╱'
		} else {
			ch = '╲'
		}
	}
	grid[y][x] = ch
}

func formatTick(v float64, unit string) string {
	if unit == "%" {
		return fmt.Sprintf("%5.1f%%", v)
	}
	return padLeft(humanBytes(v), 5)
}

func padLeft(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return strings.Repeat(" ", w-len(s)) + s
}

// renderContainerList renders a left-pane list with selection and minimal info
func renderContainerList(m *monitorModel, width, height int) string {
	var b strings.Builder
	title := boxStyle.Render("Containers")
	b.WriteString(truncateRunes(title, width))
	b.WriteByte('\n')
	count := 0
	for i, s := range m.stats {
		if count >= height-2 {
			break
		}
		line := "🐳 " + s.Name
		if i == m.selected {
			line = selStyle.Render(line)
		} else {
			line = nameStyle.Render(line)
		}
		// include quick CPU/MEM numbers on the right when space allows
		meta := fmt.Sprintf(" %3.0f%% %3.0f%%", s.CPU, s.MemPerc)
		visible := width - 1 - runeLen(meta)
		if visible < 8 {
			visible = width - 1
		}
		line = truncateRunes(line, visible)
		pad := width - 1 - runeLen(line) - runeLen(meta)
		if pad < 0 {
			pad = 0
		}
		b.WriteString(line)
		b.WriteString(strings.Repeat(" ", pad))
		if width >= 18 {
			b.WriteString(meta)
		}
		b.WriteByte('\n')
		count++
	}
	return b.String()
}

// hstackFixed puts left and right strings side-by-side with fixed left width
func hstackFixed(left, right string, leftWidth int) string {
	lLines := strings.Split(strings.TrimRight(left, "\n"), "\n")
	rLines := strings.Split(strings.TrimRight(right, "\n"), "\n")
	rows := max(len(lLines), len(rLines))
	var b strings.Builder
	for i := 0; i < rows; i++ {
		var l, r string
		if i < len(lLines) {
			l = lLines[i]
		} else {
			l = ""
		}
		if i < len(rLines) {
			r = rLines[i]
		} else {
			r = ""
		}
		if runeLen(l) < leftWidth {
			l += strings.Repeat(" ", leftWidth-runeLen(l))
		} else if runeLen(l) > leftWidth {
			l = truncateRunes(l, leftWidth)
		}
		b.WriteString(l)
		b.WriteString(" ")
		b.WriteString(r)
		b.WriteByte('\n')
	}
	return b.String()
}

func runeLen(s string) int { return len([]rune(s)) }

func truncateRunes(s string, w int) string {
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	return string(r[:w])
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// fetchOnce runs `docker stats --no-stream` and parses.
func fetchOnce() ([]DockStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--no-trunc", "--format",
		"{{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var rows []DockStats
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) < 6 {
			continue
		}
		cpu := parsePercent(cols[1])
		memp := parsePercent(cols[3])
		rows = append(rows, DockStats{
			Name:     cols[0],
			CPU:      cpu,
			MemUsage: cols[2],
			MemPerc:  memp,
			NetIO:    cols[4],
			BlockIO:  cols[5],
		})
	}
	return rows, nil
}

func parsePercent(s string) float64 {
	s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	if s == "" {
		return 0
	}
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// parseTwoBytes parses a value like "914kB / 2.25kB" into two byte totals.
func parseTwoBytes(s string) (float64, float64) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0, 0
	}
	a := parseBytes(strings.TrimSpace(parts[0]))
	b := parseBytes(strings.TrimSpace(parts[1]))
	return a, b
}

func parseBytes(s string) float64 {
	// Supports B, kB, MB, MiB, GB, GiB
	if s == "" {
		return 0
	}
	// Try simple suffixes first
	lower := strings.ToLower(s)
	var v float64
	if strings.HasSuffix(lower, "gib") {
		fmt.Sscanf(strings.TrimSuffix(lower, "gib"), "%f", &v)
		return v * 1024 * 1024 * 1024
	}
	if strings.HasSuffix(lower, "mib") {
		fmt.Sscanf(strings.TrimSuffix(lower, "mib"), "%f", &v)
		return v * 1024 * 1024
	}
	if strings.HasSuffix(lower, "kb") {
		fmt.Sscanf(strings.TrimSuffix(lower, "kb"), "%f", &v)
		return v * 1e3
	}
	if strings.HasSuffix(lower, "mb") {
		fmt.Sscanf(strings.TrimSuffix(lower, "mb"), "%f", &v)
		return v * 1e6
	}
	if strings.HasSuffix(lower, "gb") {
		fmt.Sscanf(strings.TrimSuffix(lower, "gb"), "%f", &v)
		return v * 1e9
	}
	if strings.HasSuffix(lower, "b") {
		fmt.Sscanf(strings.TrimSuffix(lower, "b"), "%f", &v)
		return v
	}
	// Fallback: attempt to read "<num><unit>" where unit may have uppercase
	unit := ""
	fmt.Sscanf(s, "%f%2s", &v, &unit)
	switch strings.ToUpper(unit) {
	case "B":
		return v
	case "KB":
		return v * 1e3
	case "MB":
		return v * 1e6
	case "MIB":
		return v * 1024 * 1024
	case "GB":
		return v * 1e9
	case "GIB":
		return v * 1024 * 1024 * 1024
	}
	return 0
}

func humanBytes(bps float64) string {
	if bps <= 0 {
		return "0B"
	}
	units := []string{"B", "kB", "MB", "GB", "TB"}
	idx := int(math.Floor(math.Log10(bps) / 3))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(units) {
		idx = len(units) - 1
	}
	scaled := bps / math.Pow(1000, float64(idx))
	if scaled >= 100 {
		return fmt.Sprintf("%d%s", int(scaled+0.5), units[idx])
	}
	return fmt.Sprintf("%.1f%s", scaled, units[idx])
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func clamp(v, lo, hi int) int {
	if hi <= 0 {
		return 0
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// NewMonitorProgram returns a function to run the Bubble Tea program.
func NewMonitorProgram() func() error {
	return func() error {
		p := tea.NewProgram(newModel(), tea.WithAltScreen())
		_, err := p.Run()
		return err
	}
}
