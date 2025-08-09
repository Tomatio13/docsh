package shell

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// replModel は Bubble Tea によるシンプルな REPL 実装
type replModel struct {
	shell        *Shell
	input        textinput.Model
	suggestions  []Suggest
	selectedIdx  int
	width        int
	height       int
	historyIndex int
	isExecuting  bool
	echoLine     string
}

func newReplModel(s *Shell) replModel {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 0
	ti.Width = 80
	// ブロックカーソル(■)は視認性が悪いので非表示
	ti.SetCursorMode(textinput.CursorHide)
	return replModel{
		shell:        s,
		input:        ti,
		suggestions:  nil,
		selectedIdx:  0,
		historyIndex: -1,
	}
}

func (m replModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tea.HideCursor)
}

type execDoneMsg struct{ err error }

func runCommandCmd(s *Shell, line string) tea.Cmd {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	return func() tea.Msg {
		err := s.executeCommand(line)
		return execDoneMsg{err: err}
	}
}

func (m replModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.Width = maxInt(20, m.width-4)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			line := m.input.Value()
			trimmed := strings.TrimSpace(line)
			if trimmed == "exit" {
				// exit はREPL側で終了
				return m, tea.Quit
			}
			if trimmed == "" {
				// 空入力: 現在のプロンプトと空入力をエコーして改行
				prompt := strings.TrimRight(m.shell.buildPrompt(), " ")
				fmt.Println(prompt)
				m.historyIndex = -1
				m.input.SetValue("")
				m.suggestions = nil
				return m, nil
			}
			// 実行中はプロンプト描画を止める
			m.isExecuting = true
			// 入力行はView側でエコーする（外部出力での重複回避）
			m.echoLine = strings.TrimRight(m.shell.buildPrompt(), " ") + " " + line
			m.input.Blur()
			m.shell.history = append(m.shell.history, line)
			m.historyIndex = -1
			m.input.SetValue("")
			m.suggestions = nil
			// login のような外部対話は REPL終了前に pendingExternal をセット
			parts := strings.Fields(trimmed)
			if len(parts) > 0 && parts[0] == "login" {
				if len(parts) < 2 {
					fmt.Println("container name required")
					m.isExecuting = false
					m.echoLine = ""
					return m, nil
				}
				if err := m.shell.enterContainer(parts[1]); err != nil {
					fmt.Printf("%v\n", err)
				}
				// コンテナから戻ったらREPLを継続
				m.isExecuting = false
				m.echoLine = ""
				m.input.Focus()
				m.input.SetCursorMode(textinput.CursorHide)
				return m, tea.HideCursor
			}
			return m, runCommandCmd(m.shell, line)
		case "tab":
			if len(m.suggestions) > 0 {
				sel := clampIndex(m.selectedIdx, len(m.suggestions))
				chosen := m.suggestions[sel].Text
				// 直近の入力がスペースで終わっている場合は新しいトークンとして追加
				line := m.input.Value()
				hasTrailingSpace := strings.HasSuffix(line, " ")
				tokens := strings.Fields(line)
				if len(tokens) == 0 {
					m.input.SetValue(chosen + " ")
				} else if hasTrailingSpace {
					// 末尾スペースあり: 既存トークンは保持し、新しいトークンを追加
					tokens = append(tokens, chosen)
					m.input.SetValue(strings.Join(tokens, " ") + " ")
				} else {
					// 末尾スペースなし: 最後のトークンを確定置換
					tokens[len(tokens)-1] = chosen
					m.input.SetValue(strings.Join(tokens, " ") + " ")
				}
				m.input.CursorEnd()
				m.suggestions = m.shell.Complete(m.input.Value())
				return m, nil
			}
		case "up":
			// サジェストがある場合は選択移動、なければ履歴
			if len(m.suggestions) > 0 {
				m.selectedIdx = (m.selectedIdx - 1 + len(m.suggestions)) % len(m.suggestions)
				return m, nil
			}
			// 履歴
			if len(m.shell.history) > 0 {
				if m.historyIndex == -1 {
					m.historyIndex = len(m.shell.history) - 1
				} else if m.historyIndex > 0 {
					m.historyIndex--
				}
				m.input.SetValue(m.shell.history[m.historyIndex])
				m.input.CursorEnd()
				m.suggestions = m.shell.Complete(m.input.Value())
				return m, nil
			}
		case "down":
			// サジェストがある場合は選択移動、なければ履歴
			if len(m.suggestions) > 0 {
				m.selectedIdx = (m.selectedIdx + 1) % len(m.suggestions)
				return m, nil
			}
			if m.historyIndex >= 0 {
				if m.historyIndex < len(m.shell.history)-1 {
					m.historyIndex++
					m.input.SetValue(m.shell.history[m.historyIndex])
				} else {
					m.historyIndex = -1
					m.input.SetValue("")
				}
				m.input.CursorEnd()
				m.suggestions = m.shell.Complete(m.input.Value())
				return m, nil
			}
		case "ctrl+n":
			if len(m.suggestions) > 0 {
				m.selectedIdx = (m.selectedIdx + 1) % len(m.suggestions)
				return m, nil
			}
		case "ctrl+p":
			if len(m.suggestions) > 0 {
				m.selectedIdx = (m.selectedIdx - 1 + len(m.suggestions)) % len(m.suggestions)
				return m, nil
			}
		}
	case execDoneMsg:
		// 実行完了後にプロンプトを復帰
		// エラーがあれば表示（従来はREPL側で非表示だったため、何も出ない問題があった）
		if msg.err != nil {
			fmt.Printf("%v\n", msg.err)
		}
		m.isExecuting = false
		m.echoLine = ""
		m.input.Focus()
		m.input.SetCursorMode(textinput.CursorHide)
		return m, tea.HideCursor
	}

	// テキスト入力側の更新
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	// 入力変更に応じてサジェスト更新
	if !m.isExecuting {
		m.suggestions = m.shell.Complete(m.input.Value())
	}
	if len(m.suggestions) == 0 {
		m.selectedIdx = 0
	} else if m.selectedIdx >= len(m.suggestions) {
		m.selectedIdx = len(m.suggestions) - 1
	}
	return m, cmd
}

var (
	promptStyle     = lipgloss.NewStyle().Bold(true)
	suggestionStyle = lipgloss.NewStyle()
	suggestionSel   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	suggestionDesc  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func (m replModel) View() string {
	var b strings.Builder
	if !m.isExecuting {
		b.WriteString(promptStyle.Render(m.shell.buildPrompt()))
		// カーソルは常に非表示
		m.input.SetCursorMode(textinput.CursorHide)
		b.WriteString(m.input.View())
		b.WriteString("\n")
	} else {
		// エコー行のみ描画（外部コマンド出力は標準出力に流れる）
		if m.echoLine != "" {
			b.WriteString(m.echoLine)
			b.WriteString("\n")
		}
	}

	// サジェスト（最大10件）
	if !m.isExecuting {
		max := minInt(10, len(m.suggestions))
		for i := 0; i < max; i++ {
			s := m.suggestions[i]
			var line string
			if i == m.selectedIdx {
				line = suggestionSel.Render(fmt.Sprintf("%s", s.Text))
			} else {
				line = suggestionStyle.Render(s.Text)
			}
			if s.Description != "" {
				line += "  " + suggestionDesc.Render(s.Description)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	return b.String()
}

// StartBubbleTeaREPL は Bubble Tea ベースの REPL を起動します
func (s *Shell) StartBubbleTeaREPL() error {
	p := tea.NewProgram(newReplModel(s))
	s.teaProgram = p
	_, err := p.Run()
	return err
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func clampIndex(v int, length int) int {
	if length <= 0 {
		return 0
	}
	if v < 0 {
		return 0
	}
	if v > length-1 {
		return length - 1
	}
	return v
}
