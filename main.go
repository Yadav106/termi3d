package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"time"

	"github.com/eiannone/keyboard"
)

type Pair struct {
	first  float64
	second float64
}

var screenHeight, screenWidth int = 40, 120
var fPlayerX, fPlayerY, fPlayerA float64 = 8.0, 8.0, 0.0

var nMapHeight, nMapWidth int = 16, 16

var fFOV float64 = -3.1415 / 4.0
var fDepth float64 = 16

var exitChan = make(chan bool)
var elapsedTime float64

var gameMap string

var showDetails bool = false

func getKeys() {
	for {
		char, _, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}

		switch string(char) {
		case "q":
			fPlayerA += 1 * elapsedTime

		case "e":
			fPlayerA -= 1 * elapsedTime

		case "a": // Move player left (adjust angle)
			strafeA := fPlayerA - math.Pi/2
			fPlayerX -= math.Sin(strafeA) * 5.0 * elapsedTime
			fPlayerY -= math.Cos(strafeA) * 5.0 * elapsedTime

			if string(gameMap[int(fPlayerY)*nMapWidth+int(fPlayerX)]) == "#" {
				fPlayerX += math.Sin(strafeA) * 5.0 * elapsedTime
				fPlayerY += math.Cos(strafeA) * 5.0 * elapsedTime
			}

		case "d": // Move player right (adjust angle)
			strafeA := fPlayerA + math.Pi/2
			fPlayerX -= math.Sin(strafeA) * 5.0 * elapsedTime
			fPlayerY -= math.Cos(strafeA) * 5.0 * elapsedTime

			if string(gameMap[int(fPlayerY)*nMapWidth+int(fPlayerX)]) == "#" {
				fPlayerX += math.Sin(strafeA) * 5.0 * elapsedTime
				fPlayerY += math.Cos(strafeA) * 5.0 * elapsedTime
			}

		case "w":
			fPlayerX += math.Sin(fPlayerA) * 5.0 * elapsedTime
			fPlayerY += math.Cos(fPlayerA) * 5.0 * elapsedTime

			if string(gameMap[int(fPlayerY)*nMapWidth+int(fPlayerX)]) == "#" {
				fPlayerX -= math.Sin(fPlayerA) * 5.0 * elapsedTime
				fPlayerY -= math.Cos(fPlayerA) * 5.0 * elapsedTime
			}

		case "s":
			fPlayerX -= math.Sin(fPlayerA) * 5.0 * elapsedTime
			fPlayerY -= math.Cos(fPlayerA) * 5.0 * elapsedTime

			if string(gameMap[int(fPlayerY)*nMapWidth+int(fPlayerX)]) == "#" {
				fPlayerX += math.Sin(fPlayerA) * 5.0 * elapsedTime
				fPlayerY += math.Cos(fPlayerA) * 5.0 * elapsedTime
			}

		case "m":
			showDetails = !showDetails

		case "c": // Exit the game
			fmt.Println("Exiting...")
			exitChan <- true // Signal the main loop to exit
			return
		}
	}
}

func main() {
	// Initialize the keyboard input
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	buffer := make([]string, screenHeight*screenWidth)

	gameMap += "################"
	gameMap += "#..............#"
	gameMap += "#..............#"
	gameMap += "#..............#"
	gameMap += "#..............#"
	gameMap += "#..........#...#"
	gameMap += "#..............#"
	gameMap += "#..........#...#"
	gameMap += "#..............#"
	gameMap += "#..............#"
	gameMap += "#......#.......#"
	gameMap += "#......#.......#"
	gameMap += "#......#########"
	gameMap += "#..............#"
	gameMap += "#..............#"
	gameMap += "################"

	fmt.Printf("\x1b[2J\x1b[?25l")
	defer fmt.Printf("\x1b[?25h")

	go getKeys()

	tp1 := time.Now()
	for true {
		tp2 := time.Now()
		elapsedTime = tp2.Sub(tp1).Seconds()
		tp1 = tp2

		select {
		case <-exitChan:
			fmt.Println("Exiting game loop...")
			return
		default:
			fmt.Printf("\x1b[H")

			for x := 0; x < screenWidth; x++ {
				fRayAngle := (fPlayerA - fFOV/2.0) + (float64(x)/float64(screenWidth))*fFOV
				fDistanceToWall := 0.0
				bHitWall := false
				bBoundary := false

				var fEyeX float64 = math.Sin(fRayAngle)
				var fEyeY float64 = math.Cos(fRayAngle)

				for !bHitWall && (fDistanceToWall < fDepth) {
					fDistanceToWall += 0.1

					var nTestX int = int(fPlayerX + fEyeX*fDistanceToWall)
					var nTestY int = int(fPlayerY + fEyeY*fDistanceToWall)

					// Test if ray is out of bounds
					if nTestX < 0 || nTestX > nMapWidth || nTestY < 0 || nTestY >= nMapHeight {
						bHitWall = true // set the distance to max depth
						fDistanceToWall = fDepth
					} else {
						// Ray is in bounds, test to see if ray cell is wall block
						if string(gameMap[nTestY*nMapWidth+nTestX]) == string("#") {
							bHitWall = true

							pairs := []Pair{}

							for tx := 0; tx < 2; tx++ {
								for ty := 0; ty < 2; ty++ {
									vy := float64(nTestY) + float64(ty) - fPlayerY
									vx := float64(nTestX) + float64(tx) - fPlayerX
									d := math.Sqrt(vx*vx + vy*vy)
									dot := (fEyeX * vx / d) + (fEyeY * vy / d)
									pairs = append(pairs, Pair{
										first:  d,
										second: dot,
									})
								}
							}

							sort.Slice(pairs, func(i, j int) bool {
								return pairs[i].first < pairs[j].first
							})

							fBound := 0.01
							if math.Acos(pairs[0].second) < fBound {
								bBoundary = true
							}
							if math.Acos(pairs[1].second) < fBound {
								bBoundary = true
							}

						}
					}
				}

				// Calculate distance to ceiling and floor
				var nCeiling int = int(float64(screenHeight/2.0) - float64(screenHeight)/float64(fDistanceToWall))
				var nFloor int = screenHeight - nCeiling

				var shade rune = ' '

				if fDistanceToWall <= fDepth/4.0 {
					shade = 0x2588
				} else if fDistanceToWall <= fDepth/3.0 {
					shade = 0x2593
				} else if fDistanceToWall <= fDepth/2.0 {
					shade = 0x2592
				} else if fDistanceToWall <= fDepth {
					shade = 0x2591
				} else {
					shade = ' '
				}

				if bBoundary {
					shade = rune(' ')
				}

				for y := 0; y < screenHeight; y++ {
					if y < nCeiling {
						buffer[y*screenWidth+x] = " "
					} else if y > nCeiling && y < nFloor {
						buffer[y*screenWidth+x] = string(shade)
					} else {
						b := 1.0 - ((float64(y) - float64(screenHeight)/2.0) / (float64(screenHeight) / 2.0))
						fShade := " "
						if b < 0.25 {
							fShade = "#"
						} else if b < 0.5 {
							fShade = "x"
						} else if b < 0.75 {
							fShade = "."
						} else if b < 0.9 {
							fShade = "-"
						} else {
							fShade = " "
						}
						buffer[y*screenWidth+x] = fShade
					}
				}

			}

			if showDetails {
				// Display Stats
				formattedString := fmt.Sprintf("X=%3.2f, Y=%3.2f, A=%3.2f FPS=%3.2f", fPlayerX, fPlayerY, fPlayerA, 1.0/elapsedTime)
				for i, char := range formattedString {
					if i < screenWidth { // Ensure we don't exceed the screen width
						buffer[i] = string(char)
					}
				}

				// Display Map
				for nx := 0; nx < nMapWidth; nx++ {
					for ny := 0; ny < nMapWidth; ny++ {
						buffer[(ny+1)*screenWidth+nx] = string(gameMap[ny*nMapWidth+nx])
					}
				}

				// Display Player
				buffer[(int(fPlayerY)+1)*screenWidth+int(fPlayerX)] = "P"
			}

			// Output the buffer to the screen
			for y := 0; y < screenHeight; y++ {
				for x := 0; x < screenWidth; x++ {
					fmt.Printf("%s", buffer[y*screenWidth+x])
				}
				fmt.Println()
			}

			time.Sleep(time.Microsecond * 8000 * 2)
		}
	}
}

func getLogger(fileName string) *log.Logger {
	logfile, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic("give a better file 🗿")
	}

	return log.New(logfile, "[ascii]", log.Ldate|log.Ltime)
}
