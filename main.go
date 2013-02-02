package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/nsf/termbox-go"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	x, y            int
	sx, sy          int
	tx, ty          int
	gtx, gty        int
	last            int
	page            int
	array           []block
	input, oldinput []rune
	bookmarks       = os.Getenv("HOME") + "/.conkeror.mozdev.org/bookmarks.json"
)

type block struct {
	b    bm
	x, y int
}

type bm struct {
	Name string
	URL  string
	Tags []string
}

func main() {
	if len(os.Args) >= 3 {
		var newbm bm
		for k, v := range os.Args {
			switch k {
			case 0:
				break
			case 1:
				newbm.Name = v
			case 2:
				newbm.URL = v
			default:
				for _, vv := range strings.Split(v, " ") {
					newbm.Tags = append(newbm.Tags, vv)
				}
			}
		}
		add(&newbm)
		return
	}
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputAlt)
	defer termbox.Close()
	open()
	redraw()
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyPgup:
				if page != 0 && y >= 0 {
					y = -1
				} else {
					y = 0
				}				
			case termbox.KeyCtrlV, termbox.KeyPgdn:
				y = array[last].y
			case termbox.KeyEsc, termbox.KeyCtrlC, termbox.KeyCtrlQ:
				return
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				if len(input) != 0 {
					input = input[:len(input)-1]
				}
			case termbox.KeyEnter:
				for _, v := range array {
					if v.y == y {
						termbox.Close()
						exec.Command("/usr/bin/conkeror", v.b.URL).Start()
						time.Sleep(300 * time.Millisecond)
						os.Exit(0)
					}
				}
			case termbox.KeyArrowUp, termbox.KeyCtrlP:
				if !(page == 0 && y == 0) {
					y = y - 2
				}
			case termbox.KeyArrowDown, termbox.KeyCtrlN:
				if y < array[last].y {
					y = y + 2
				}
			case termbox.KeyCtrlU:
				input = append(input, []rune("URL:.*")...)
			case termbox.KeyCtrlT:
				input = append(input, []rune("Tags:.*")...)
			case termbox.KeySpace:
				input = append(input, ' ')
			case termbox.KeyCtrlG:
				input = []rune("")
			}
			if ev.Ch != 0 {
				if ev.Mod == termbox.ModAlt && ev.Ch == 118 {
					if page != 0 && y >= 0 {
						y = -1
					} else {
						y = 0
					}
				} else {
					input = append(input, ev.Ch)
				}
			}
			fallthrough
		default:
			redraw()
		}
	}
}

func redraw() {
	tx, ty = 0, 0
	sx, sy = termbox.Size()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if string(input) != string(oldinput) {
		gty = 0
		y = 0
	}
	for k, v := range array {
		if ok, _ := regexp.MatchString(string(input), fmt.Sprintf("%+v", v.b)); ok {
			if y > sy {
				gty = gty - sy
				y = 0
				page++
			} else if y < 0 {
				gty = gty + sy
				y = sy - 2
				page--
			}
			array[k].x, array[k].y = tx, ty+gty
			tx, ty = tx+2, ty+2
			printf(0, array[k].y, termbox.ColorDefault, termbox.ColorDefault, " %v %v", v.b.Name, v.b.Tags)
			printf(0, array[k].y+1, termbox.ColorWhite, termbox.ColorDefault, "    %v", v.b.URL)
			last = k
		} else {
			array[k].y = -10
		}
		printf(0, sy-1, termbox.ColorGreen, termbox.ColorDefault, "%v", string(input))
	}
	oldinput = input
	termbox.SetCursor(x, y)
	termbox.Flush()
}

func open() {
	file, err := os.OpenFile(bookmarks, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		str, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		var bl block
		json.Unmarshal(str, &bl.b)
		array = append(array, bl)
	}
}

func add(newbm *bm) {
	s, err := json.Marshal(newbm)
	if err != nil {
		fmt.Println(err)
		return
	}
	file, err := os.OpenFile(bookmarks, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	if _, err := file.Write(s); err != nil {
		fmt.Println(err)
		return
	}
	file.WriteString("\n")
}

func printf(x, y int, fg, bg termbox.Attribute, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	for _, c := range s {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}