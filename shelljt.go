package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"syscall"

	gc "github.com/rthornton128/goncurses"
	"flag"
)

type state struct {
	matches  map[int]*section
	content  string
	selected string
	index    map[rune]*section
}

type section struct {
	begin int
	end   int
}

func main() {
	filename := flag.String("outputFilename", "output file", "File to output selected contents to")
	flag.Parse()

	//in := bufio.NewReader(os.Stdin)
	stats, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("file.Stat()", err)
		return
	}

	if stats.Size() < 0 {
		fmt.Println("Nothing on STDIN")
		return
	}

	b, err := ioutil.ReadAll(os.Stdin)

	if err != nil {
		fmt.Println("STDIO error", err)
		return
	}

	f, err := os.Open("/dev/tty")
	if err != nil {
		fmt.Println("Error reopening /dev/tty", err)
		return
	}
	syscall.Dup2(int(f.Fd()), 0)

	stdscr, _ := gc.Init()
	defer gc.End()

	// Read text, if regex match replace first character (which is highlighted)
	// write to screen
	// have an input to type 'a-z'
	// fully highlight match
	// enter prints all matches to term

	s := state{
		content:  string(b),
		selected: "",
		matches:  map[int]*section{},
		index:    map[rune]*section{},
	}

	r := regexp.MustCompile(`(?m)\S+`)
	indexes := r.FindAllStringIndex(string(b), -1)
	ascii := 'a'
	for index := range indexes {
		sc := section{begin: indexes[index][0], end: indexes[index][1]}
		s.matches[sc.begin] = &sc
		s.index[ascii] = &sc
		ascii++
	}

	update(&s, stdscr)

	var ch gc.Key
	for ch != 'q' {
		ch = stdscr.GetChar()

		if ch == gc.KEY_RETURN {
			gc.End()
			b := ""
			for _, c := range s.selected {
				b += s.content[s.index[c].begin:s.index[c].end]
			}
			b += "\n"
			ioutil.WriteFile(*filename, []byte(b), 0644)
			return
		}

		if strings.Contains(s.selected, string(ch)) {
			// Remove
			s.selected = strings.Replace(s.selected, string(ch), "", -1)
		} else {
			// Add
			s.selected += string(ch)
		}
		update(&s, stdscr)
	}
}

func update(s *state, stdscr *gc.Window) {
	stdscr.Clear()
	var last_found *section
	var ascii rune
	for i := 0; i < len(s.content); i++ {
		found := s.matches[i]
		if found != nil {
			last_found = found
			// Print the associated key instead of the real letter
			stdscr.AttrOn(gc.A_BOLD)
			for k, v := range s.index {
				if found == v {
					ascii = k
					break
				}
			}
			stdscr.Print(string(ascii))
			stdscr.AttrOff(gc.A_BOLD)
		} else {
			// Print the real letter, in bold if selected
			if last_found != nil && i >= last_found.begin && i <= last_found.end && strings.Contains(s.selected, string(ascii)) {
				stdscr.AttrOn(gc.A_BOLD)
			}
			stdscr.Print(string(s.content[i]))
			stdscr.AttrOff(gc.A_BOLD)
		}
	}

	//	row, _ := stdscr.MaxYX()
	//	stdscr.MovePrint(row - 1, 0, "selected = " + s.selected + "] >> ")
	stdscr.Refresh()
}
