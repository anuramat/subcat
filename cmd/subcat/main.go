package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

type (
	fileCloseMsg  int
	fileUpdateMsg struct {
		line    string
		filenum int
	}
	file struct {
		filename string
		filenum  int
		lines    []string
		isReg    bool // read on inotify, otherwise read until EOF
		changed  chan struct{}
	}
)

func makeFiles(filenames []string, w *fsnotify.Watcher) ([]file, error) {
	files := []file{}
	for i, name := range filenames {
		stat, err := os.Stat(name)
		if err != nil {
			return nil, err
		}
		f := file{
			filename: name,
			filenum:  i,
			lines:    []string{},
			isReg:    stat.Mode().IsRegular(),
		}
		if f.isReg {
			w.Add(f.filename)
			f.changed = make(chan struct{})
		}
		files = append(files, f)
	}
	return files, nil
}

type model struct {
	files []file
}

func newModel(files []file) model {
	m := model{files: files}
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
	case fileUpdateMsg:
		m.files[msg.filenum].lines = append(m.files[msg.filenum].lines, msg.line)
	}
	return m, nil
}

func (m model) View() string {
	content := []string{}
	for _, file := range m.files {
		content = append(content, strings.Join(file.lines, ""))
	}
	return strings.Join(content, "")
}

func processFile(p *tea.Program, f file, wg *sync.WaitGroup) {
	defer wg.Done()
	defer f.close(p)

	file, err := os.Open(f.filename)
	if err != nil {
		log.Fatalf("Error opening file %s: %v\n", f.filename, err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Fatalf("Error reading file %s: %v\n", f.filename, err)
			} else if !f.isReg {
				break
			} else if f.isReg {
				<-f.changed
			}
		}
		f.write(p, line)
	}
}

func watch(w *fsnotify.Watcher, chans map[string]chan struct{}) error {
	for {
		select {
		case err, ok := <-w.Errors:
			if !ok {
				return err
			}
		case e, ok := <-w.Events:
			if !ok {
				return nil
			}
			chans[e.Name] <- struct{}{}
		}
	}
}

func (r file) close(p *tea.Program) {
	p.Send(fileCloseMsg(r.filenum))
}

func (r file) write(p *tea.Program, line string) {
	p.Send(fileUpdateMsg{
		line:    line,
		filenum: r.filenum,
	})
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln("Error: no arguments")
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Can't make a watcher: %v\n", err)
	}
	defer w.Close()

	files, err := makeFiles(os.Args[1:], w)
	if err != nil {
		log.Fatalf("File init err: %v\n", err)
	}
	chanMap := make(map[string]chan struct{})
	for _, f := range files {
		chanMap[f.filename] = f.changed
	}
	go watch(w, chanMap)

	m := newModel(files)
	p := tea.NewProgram(m)

	wg := &sync.WaitGroup{}
	for _, file := range files {
		wg.Add(1)
		go processFile(p, file, wg)
	}
	go func() {
		wg.Wait()
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		log.Fatalf("Bubbletea error: %v\n", err)
		os.Exit(1)
	}
}
