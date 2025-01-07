package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	fileClose  struct{}
	fileUpdate struct {
		line    string
		filenum int
	}
)

type model struct {
	files   [][]string
	counter int
}

func initialModel(n int) model {
	m := model{}
	for range n {
		m.files = append(m.files, []string{})
		m.counter = n
	}
	return m
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("subcat")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case fileUpdate:
		m.files[msg.filenum] = append(m.files[msg.filenum], msg.line)
	case fileClose:
		m.counter--
	}
	if m.counter == 0 {
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	files := []string{}
	for _, lines := range m.files {
		files = append(files, strings.Join(lines, "\n")+"\n")
	}
	return strings.Join(files, "")
}

func processFile(filename string, wg *sync.WaitGroup, write func(string), close func()) {
	defer wg.Done()
	defer close()

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening file %s: %v\n", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t := scanner.Text()
		write(t)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file %s: %v\n", filename, err)
	}
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln("Error: no arguments")
	}

	p := tea.NewProgram(initialModel(len(os.Args) - 1))
	wg := sync.WaitGroup{}

	for i, filename := range os.Args[1:] {
		wg.Add(1)
		go processFile(filename, &wg, (func(s string) {
			p.Send(fileUpdate{
				line:    s,
				filenum: i,
			})
		}),
			func() { p.Send(fileClose{}) },
		)
	}

	if _, err := p.Run(); err != nil {
		log.Fatalf("Bubbletea error: %v\n", err)
		os.Exit(1)
	}
}
