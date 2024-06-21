package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

type todoMenu struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

type addMenu struct {
	text           []rune
	textLength     int
	cursorPosition int
}

type activeMenu = int

const todoActive activeMenu = 0
const addActive activeMenu = 1

const HEADER_LEN int = 35

type model struct {
	menuActive activeMenu
	todoStruct *todoMenu
	addStruct  *addMenu
	filename   string
}

func InitialTodoMenu() *todoMenu {
	return &todoMenu{
		selected: make(map[int]struct{}),
	}
}

func InitialModel() model {
	return model{
		menuActive: todoActive,
		todoStruct: InitialTodoMenu(),
		addStruct:  &addMenu{text: []rune{}, textLength: 0},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (t *todoMenu) handleKey(key string) {
	switch key {
	case "up", "k":
		if t.cursor > 0 {
			t.cursor--
		}

	case "down", "j":
		if t.cursor < len(t.choices)-1 {
			t.cursor++
		}

	case "enter", " ":
		_, ok := t.selected[t.cursor]
		if ok {
			delete(t.selected, t.cursor)
		} else {
			t.selected[t.cursor] = struct{}{}
		}

	case "delete":
		if len(t.choices) == 1 {
			t.choices = []string{}
			delete(t.selected, 0)
			return
		}
		new_todos := make([]string, len(t.choices)-1)
		delete(t.selected, t.cursor)
		for i := 0; i < len(t.choices); i++ {
			if i < t.cursor {
				new_todos[i] = t.choices[i]
			}
			if i > t.cursor {
				new_todos[i-1] = t.choices[i]
				if _, ok := t.selected[i]; ok {
					t.selected[i-1] = struct{}{}
				}
				delete(t.selected, i)
			}
		}
		t.choices = new_todos
		t.cursor = max(t.cursor-1, 0)
	}
}

func (a *addMenu) handleKey(key string, m *todoMenu) {
	switch key {
	case "left":
		if a.cursorPosition > 0 {
			a.cursorPosition--
		}
	case "right":
		if a.cursorPosition < a.textLength {
			a.cursorPosition++
		}
	case "enter":
		if len(a.text) == 0 {
			return
		}
		m.choices = append(m.choices, strings.TrimRight(string(a.text), " "))
		a.text = []rune{}
		a.textLength = 0
		a.cursorPosition = 0
	case "backspace", "delete":
		if len(a.text) == 0 {
			return
		}
		if a.cursorPosition == a.textLength {
			a.text = a.text[:a.textLength-1]
			a.textLength--
			a.cursorPosition--
			return
		}
		before := a.text[:a.cursorPosition]
		after := a.text[a.cursorPosition+1:]
		a.text = append(before, after...)
		a.textLength--
		if a.cursorPosition > 0 {
			a.cursorPosition--
		}
	default:
		runes := []rune(key)
		if len(runes) > 1 {
			return
		}
		first_char := runes[0]
		if !unicode.IsGraphic(first_char) {
			return
		}
		before := make([]rune, a.cursorPosition)
		copy(before, a.text[:a.cursorPosition])
		after := a.text[a.cursorPosition:]
		before = append(before, first_char)
		before = append(before, after...)
		a.text = before
		a.cursorPosition++
		a.textLength++
	}
}

func (m *model) HandleKey(key string) {
	if m.menuActive == todoActive {
		m.todoStruct.handleKey(key)
	} else {
		m.addStruct.handleKey(key, m.todoStruct)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			m.writeLines()
			return m, tea.Quit
		case "tab":
			m.menuActive ^= 1
			return m, nil
		default:
			m.HandleKey(msg.String())
		}
	}
	return m, nil
}

func formatText(text string, highlight bool) string {
	if highlight {
		return fmt.Sprintf("%s     %s     %s", "\x1b[7m", text, "\x1b[27m")
	}
	return fmt.Sprintf("     %s     ", text)
}

func (m model) View() string {
	var str strings.Builder
	str.WriteString(formatText("Todos", m.menuActive == todoActive))
	str.WriteByte('|')
	str.WriteString(formatText("Add Todos", m.menuActive == addActive) + "\n")
	str.WriteString(strings.Repeat("-", HEADER_LEN) + "\n\n")
	if m.menuActive == todoActive {
		str.WriteString("What should I do?\n\n")

		for i, choice := range m.todoStruct.choices {
			cursor := " "
			if m.todoStruct.cursor == i {
				cursor = ">"
			}

			checked := " "
			if _, ok := m.todoStruct.selected[i]; ok {
				checked = "x"
			}

			fmt.Fprintf(&str, "%s [%s] %s\n", cursor, checked, choice)
		}
		if len(m.todoStruct.choices) == 0 {
			str.WriteString("There is currently nothing to do!!\n")
		}
	} else {
		str.WriteString("Add a new todo:\n\n" + "Input: ")
		str.WriteString(string(m.addStruct.text))
		fmt.Fprintf(&str, "\n%s^\n", strings.Repeat(" ", len("Input: ")+m.addStruct.cursorPosition))
		if len(m.addStruct.text) > 0 {
			str.WriteString("Press `enter` to add todo\n")
		}
	}
	str.WriteString("\nPress `ctrl+q` to quit\n")
	return str.String()
}

func (m *model) getLines() {
	file, err := os.Open(m.filename)
	if err != nil {
		fmt.Printf("There was an error opening the file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	result := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	m.todoStruct.choices = result
}

func (m *model) writeLines() {
	file, err := os.Create(m.filename + "_tmp")
	if err != nil {
		fmt.Printf("Error while writing to the file: %v", err)
		os.Exit(1)
	}
	defer file.Close()
	for _, line := range m.todoStruct.choices {
		fmt.Fprintln(file, line)
	}
	err = os.Rename(m.filename+"_tmp", m.filename)
	if err != nil {
		fmt.Printf("Error while writing to the file: %v", err)
		os.Exit(1)
	}
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Invalid usage of app")
		fmt.Println("Try: `./go-todo-tui \"file to load todos from\"`")
		os.Exit(1)
	}
	filename := args[0]
	model := InitialModel()
	model.filename = filename
	model.getLines()
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("There has been an error: %v", err)
		os.Exit(1)
	}
}
