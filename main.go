package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/eiannone/keyboard"
)

var screenHeight, screenWidth int = 40, 120
var fPlayerX, fPlayerY, fPlayerA float64 = 8.0, 8.0, 0.0

var nMapHeight, nMapWidth int = 16, 16

var fFOV float64 = 3.1415 / 4.0
var fDepth float64 = 16

var exitChan = make(chan bool)
var elapsedTime float64

func getKeys() {
	for {
		char, _, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}

		switch string(char) {
		case "a": // Move player left (adjust angle)
			fPlayerA -= 0.8 * elapsedTime
		case "d": // Move player right (adjust angle)
			fPlayerA += 0.8 * elapsedTime
    case "w":
      fPlayerX += math.Sin(fPlayerA) * 5.0 * elapsedTime
      fPlayerY += math.Cos(fPlayerA) * 5.0 * elapsedTime
    case "s":
      fPlayerX -= math.Sin(fPlayerA) * 5.0 * elapsedTime
      fPlayerY -= math.Cos(fPlayerA) * 5.0 * elapsedTime
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

	var gameMap string
	gameMap += "################"
	gameMap += "#..............#"
	gameMap += "#..............#"
	gameMap += "#..............#"
	gameMap += "#..............#"
	gameMap += "#..........#...#"
	gameMap += "#..........#...#"
	gameMap += "#..............#"
	gameMap += "#..............#"
	gameMap += "#..............#"
	gameMap += "#..............#"
	gameMap += "#..............#"
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
						}
					}
				}

				// Calculate distance to ceiling and floor
				var nCeiling int = int(float64(screenHeight/2.0) - float64(screenHeight)/float64(fDistanceToWall))
				var nFloor int = screenHeight - nCeiling

        var shade rune = ' '

        if fDistanceToWall <= fDepth / 4.0 {
          shade = 0x2588
        } else if fDistanceToWall <= fDepth / 3.0 {
          shade = 0x2593
        } else if fDistanceToWall <= fDepth / 2.0 {
          shade = 0x2592
        } else if fDistanceToWall <= fDepth {
          shade = 0x2591
        } else {
          shade = ' '
        }

				for y := 0; y < screenHeight; y++ {
					if y < nCeiling {
						buffer[y*screenWidth+x] = " "
					} else if y > nCeiling && y < nFloor {
						buffer[y*screenWidth+x] = string(shade)
					} else {
						buffer[y*screenWidth+x] = " "
					}
				}

			}

			// Output the buffer to the screen
			for y := 0; y < screenHeight; y++ {
				for x := 0; x < screenWidth; x++ {
					fmt.Printf("%s", buffer[y*screenWidth+x])
				}
				fmt.Println()
			}

			// fPlayerA -= 0.1

			time.Sleep(time.Microsecond * 8000 * 2)
		}
	}
}
