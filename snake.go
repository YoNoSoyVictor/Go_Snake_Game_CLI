package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// SNAKE GAME FOR CLI. DISPLAYS CORRECTLY IN BASH, IF YOU WANT TO USE A DIFFERENT SHELL CHECK THE CONVENIENT
// ESCAPE SEQUENCES FOR SAVING AND RESTORING CURSOR POSITION, AND EDIT IT IN THE printBoard FUNCTION

// THIS PROGRAM USES EMOJIS AND POWERLINE SYMBOLS BY DEFAULT. IF YOUR TERMINAL DOES NOT SUPPORT EMOJIS OR
// POWERLINE SYMBOLS YOU CAN CHANGE IT BELOW TO SOMETHING THAT CAN BE CORRECTLY DESPLAYED IN YOUR TERMINAL

// ##### CUSTOMIZATION #####
// CONTROLS
const upKey string = "w"
const leftKey string = "a"
const downKey string = "s"
const rightKey string = "d"
const quitKey string = "q"

// APPEARANCE
const topLimL string = "╔"
const topLimR string = "╗"
const downLimL string = "╚"
const downLimR string = "╝"
const vertLim string = "║"
const horzLim string = "══"
const void string = "  "

const snakebody string = "██"
const food string = ""

const topmessage string = "🐍SNAKEGAME. MOVE WITH " + upKey + ", " + leftKey + ", " + downKey + ", " + rightKey + "."
const bottommessage string = "PRESS " + quitKey + " TO QUIT."

func main() {

	//flagging
	var length int
	var height int
	var speed int
	flag.IntVar(&length, "length", 16, "Length of the board.")
	flag.IntVar(&height, "heigth", 12, "Heigth of the board.")
	flag.IntVar(&speed, "speed", 100, "Speed of the game, in milliseconds.")
	flag.Parse()

	//initializing elements and variables
	sleeptime := time.Duration(speed)
	totaltiles := length * height

	starting_position := height * length / 2
	snake := snake{coordinates: []int{starting_position}}
	snake.atefruit = true

	myboard := board{height: height, length: length}
	myboard.createBoard()
	myboard.calculateOccupiedTiles()

	gamechannel := make(chan string)

	//hide inputs
	hideInput()
	defer showInput()

	//wait for input to start game

	printBoard(myboard.layout, nil, length)
	snake.currentdirection = readkeyNoChannel()

gameloop:
	for {
		go readkey(gamechannel)

		printBoard(myboard.layout, snake.coordinates, length)

		time.Sleep(sleeptime * time.Millisecond)

		select {
		case key := <-gamechannel:
			if key == leftKey {
				if snake.currentdirection == rightKey {
					snake.moveRight()
				} else {
					snake.moveLeft()
					snake.currentdirection = leftKey
				}

			} else if key == downKey {
				if snake.currentdirection == upKey {
					snake.moveUp(length)
				} else {
					snake.moveDown(length)
					snake.currentdirection = downKey
				}
			} else if key == rightKey {
				if snake.currentdirection == leftKey {
					snake.moveLeft()
				} else {
					snake.moveRight()
					snake.currentdirection = rightKey
				}
			} else if key == upKey {
				if snake.currentdirection == downKey {
					snake.moveDown(length)
				} else {
					snake.moveUp(length)
					snake.currentdirection = upKey
				}
			} else if key == quitKey {
				break gameloop
			}

		default:
			if snake.currentdirection == leftKey {
				snake.moveLeft()
			} else if snake.currentdirection == downKey {
				snake.moveDown(length)
			} else if snake.currentdirection == rightKey {
				snake.moveRight()
			} else if snake.currentdirection == upKey {
				snake.moveUp(length)
			}
		}

		snake.headcoordinate = snake.coordinates[len(snake.coordinates)-1]

		if snake.atefruit == true {
			myboard.layout[getRandIntWithExclusion(length, totaltiles, append(myboard.occupiedTiles, snake.coordinates...))] = food
		} else {
			snake.loseTail()
		}
		snake.atefruit = false

		if myboard.layout[snake.headcoordinate] == vertLim || myboard.layout[snake.headcoordinate] == horzLim {
			break
		}

		if checkDuplicateInt(snake.coordinates) == true {
			break
		}

		if myboard.layout[snake.headcoordinate] == food {
			snake.atefruit = true
			myboard.layout[snake.headcoordinate] = void
		}

	}

	gameOverScreen(height, len(snake.coordinates))
}

type snake struct {
	coordinates      []int // the last value is the head.
	atefruit         bool
	headcoordinate   int
	currentdirection string
}

type board struct {
	layout        []string // layout of the board.
	occupiedTiles []int    // holds the coordinates of the tiles where the snake can't step. Faster to parse than layout.
	height        int
	length        int
}

// ##### SNAKE MOVEMENT #####

func (s *snake) moveLeft() {
	s.coordinates = append(s.coordinates, s.coordinates[len(s.coordinates)-1]-1)
}

func (s *snake) moveRight() {
	s.coordinates = append(s.coordinates, s.coordinates[len(s.coordinates)-1]+1)
}

//both moveup and movedown require the len of a row as an argument
func (s *snake) moveUp(l int) {
	s.coordinates = append(s.coordinates, s.coordinates[len(s.coordinates)-1]-l-2)
}

func (s *snake) moveDown(l int) {
	s.coordinates = append(s.coordinates, s.coordinates[len(s.coordinates)-1]+l+2)
}

func (s *snake) loseTail() {
	s.coordinates = s.coordinates[1:]
}

// ##### BOARD #####

func (b *board) createBoard() {

	//creates the board layout where the snake will move, visual purposes.
	var height int = b.height + 2
	var length = b.length + 2

	board := make([]string, height*length)

	for i := 0; i < height*length; i++ {
		board[i] = void
	}

	for i := 0; i < height*length; i++ {
		if i%length == 0 {
			board[i] = vertLim
			board[i+length-1] = vertLim
		}
	}

	for i := 0; i < length; i++ {
		board[i] = horzLim
	}

	for i := height*length - length; i < height*length; i++ {
		board[i] = horzLim
	}

	board[0] = topLimL
	board[length-1] = topLimR
	board[height*length-length] = downLimL
	board[height*length-1] = downLimR

	b.layout = board
}

func (b *board) calculateOccupiedTiles() {

	//creates occupied tiles slice using the board
	for i, elem := range b.layout {
		if elem != void {
			b.occupiedTiles = append(b.occupiedTiles, i)
		}
	}
}

// ##### DISPLAY/GAME #####

func printBoard(board []string, snake_coord []int, len int) {

	//Puts the snake over the board and formats it nicely to display.
	var screen string

	for i, tile := range board {

		if i%(len+2) == 0 {
			screen += "\n"
		}

		var flag bool = false
		for _, value := range snake_coord {
			if i == value {
				flag = true
			}

		}
		if flag == true {
			screen += snakebody
		} else {
			screen += tile
		}

	}

	//fmt.Printf("\033[s Prueba %s \033[u", str)
	fmt.Printf("\033[s %s%s\n%s \033[u", topmessage, screen, bottommessage)
}

func gameOverScreen(height int, snake_len int) {

	//Displays a game over screen.
	newpromptheight := strconv.Itoa(height + 3)
	finalscore := strconv.Itoa(snake_len)
	fmt.Printf("\033[%sBGame Over! your final score is %s\n", newpromptheight, finalscore)
}

// ##### INPUT/TERMINAL #####

func hideInput() {

	// disable input buffering, A.K.A. having to press enter.
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
}

func showInput() {

	// restore showing entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "echo").Run()
}

func readkey(c chan string) {

	//	for {
	//		var b = make([]byte, 1)
	//		os.Stdin.Read(b)
	//		c <- string(b)
	//	}

	var b = make([]byte, 1)
	os.Stdin.Read(b)
	var value string = string(b)
	if value == upKey || value == leftKey || value == downKey || value == rightKey || value == quitKey {
		c <- value
	}
}

func readkeyNoChannel() string {

	var b = make([]byte, 1)
	os.Stdin.Read(b)
	var value string = string(b)
	if value == upKey || value == leftKey || value == downKey || value == rightKey || value == quitKey {
		return value
	}
	return upKey
}

// ##### HELPER FUNCTIONS #####

func checkDuplicateInt(intSlice []int) bool {

	for x, item := range intSlice {
		for y, value := range intSlice {
			if item == value && x != y {
				return true
			}
		}
	}
	return false
}

func getRandIntWithExclusion(min, max int, blacklisted []int) int {
	// if blacklisted is/can be large, you might want to think about caching it
	excluded := map[int]bool{}
	for _, x := range blacklisted {
		excluded[x] = true
	}

	// loop until an n is generated that is not in the blacklist
	for {
		n := min + rand.Intn(max+1) // yields n such that min <= n <= max
		if !excluded[n] {
			return n
		}
	}
}
