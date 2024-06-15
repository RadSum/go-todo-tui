package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

type todo_menu struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

type add_menu struct {
	text []rune
    text_length int
}

type active_menu = int
const todo_active active_menu = 0
const add_active active_menu = 1

const HEADER_LEN int = 35 

type model struct {
	menu_active active_menu
	todo_struct *todo_menu
	add_struct  *add_menu
}

func InitialTodoMenu(initial_todos []string) *todo_menu {
	return &todo_menu{
		choices:  initial_todos,
		selected: make(map[int]struct{}),
	}
}

func InitialModel(initial_todos []string) model {
	return model{
		menu_active: todo_active,
		todo_struct: InitialTodoMenu(initial_todos),
        add_struct:  &add_menu{text: []rune{}, text_length: len("Input: ")},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) HandleKey(key string) {
	if m.menu_active == todo_active {
		t := m.todo_struct
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
		}
	} else {
        if (key == "enter" && len(m.add_struct.text) > 0) {
            m.todo_struct.choices = append(m.todo_struct.choices, string(m.add_struct.text))
            m.add_struct.text = []rune{}
            m.add_struct.text_length = len("Input: ")
        }

		if (key == "backspace" || key == "delete") && len(m.add_struct.text) > 0 {
			m.add_struct.text = m.add_struct.text[:len(m.add_struct.text)-1]
            m.add_struct.text_length--
		}
        runes := []rune(key)
		if len(runes) > 1 {
			return
		}
        first_char := runes[0] 
        if !unicode.IsGraphic(first_char)  {
            return 
        }
		m.add_struct.text = append(m.add_struct.text, first_char)
        m.add_struct.text_length++
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "tab":
			m.menu_active ^= 1
			return m, nil
		default:
			m.HandleKey(msg.String())
		}
	}
	return m, nil
}

func format_text(text string, highlight bool) string {
    if highlight {
        return fmt.Sprintf("%s     %s     %s", "\x1b[7m", text, "\x1b[27m")
    }
    return fmt.Sprintf("     %s     ", text)
}

func (m model) View() string {
	s := ""
    s += format_text("Todos", m.menu_active == todo_active)
    s += "|"
    s += format_text("Add Todos", m.menu_active == add_active) + "\n"
    s += strings.Repeat("-", HEADER_LEN)
    s += "\n\n"
	if m.menu_active == todo_active {
		s += "What should we buy at the market\n\n"

		for i, choice := range m.todo_struct.choices {
			cursor := " "
			if m.todo_struct.cursor == i {
				cursor = ">"
			}

			checked := " "
			if _, ok := m.todo_struct.selected[i]; ok {
				checked = "x"
			}

			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
		}
	} else {
		s += "Add a new todo:\n\n"
		s += "Input: "
		s += string(m.add_struct.text)
		s += "\n"
		s += strings.Repeat(" ", m.add_struct.text_length)
		s += "^\n"
        if len(m.add_struct.text) > 0 {
            s += "Press `enter` to add todo\n"
        }
	}
	s += "\nPress `ctrl+q` to quit\n"
	return s
}

func main() {
	p := tea.NewProgram(InitialModel([]string{}))
	if _, err := p.Run(); err != nil {
		fmt.Printf("There has been an error: %v", err)
		os.Exit(1)
	}
}
