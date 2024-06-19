package main

import (
	"bufio"
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
	text        []rune
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
	filename    string
}

func InitialTodoMenu() *todo_menu {
	return &todo_menu{
		selected: make(map[int]struct{}),
	}
}

func InitialModel() model {
	return model{
		menu_active: todo_active,
		todo_struct: InitialTodoMenu(),
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
	} else {
		if key == "enter" && len(m.add_struct.text) > 0 {
			m.todo_struct.choices = append(m.todo_struct.choices, strings.TrimRight(string(m.add_struct.text), " "))
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
		if !unicode.IsGraphic(first_char) {
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
			m.write_lines()
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
	var str strings.Builder
	str.WriteString(format_text("Todos", m.menu_active == todo_active))
	str.WriteByte('|')
	str.WriteString(format_text("Add Todos", m.menu_active == add_active) + "\n")
	str.WriteString(strings.Repeat("-", HEADER_LEN) + "\n\n")
	if m.menu_active == todo_active {
		str.WriteString("What should I do?\n\n")

		for i, choice := range m.todo_struct.choices {
			cursor := " "
			if m.todo_struct.cursor == i {
				cursor = ">"
			}

			checked := " "
			if _, ok := m.todo_struct.selected[i]; ok {
				checked = "x"
			}

			fmt.Fprintf(&str, "%s [%s] %s\n", cursor, checked, choice)
		}
		if len(m.todo_struct.choices) == 0 {
			str.WriteString("There is currently nothing to do!!\n")
		}
	} else {
		str.WriteString("Add a new todo:\n\n" + "Input: ")
		str.WriteString(string(m.add_struct.text))
		fmt.Fprintf(&str, "\n%s^\n", strings.Repeat(" ", m.add_struct.text_length))
		if len(m.add_struct.text) > 0 {
			str.WriteString("Press `enter` to add todo\n")
		}
	}
	str.WriteString("\nPress `ctrl+q` to quit\n")
	return str.String()
}

func (m *model) get_lines() {
	file, err := os.Open(m.filename)
	defer file.Close()
	if err != nil {
		fmt.Printf("There was an error opening the file: %v\n", err)
		os.Exit(1)
	}
	result := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	m.todo_struct.choices = result
}

func (m *model) write_lines() {
	file, err := os.Create(m.filename + "_tmp")
	defer file.Close()
	if err != nil {
		fmt.Printf("Error while writing to the file: %v", err)
		os.Exit(1)
	}
	for _, line := range m.todo_struct.choices {
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
	model.get_lines()
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("There has been an error: %v", err)
		os.Exit(1)
	}
}
