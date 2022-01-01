package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/AlexdeRuijter/stochastic-particle-model/filepool"
	"github.com/AlexdeRuijter/stochastic-particle-model/schemes"
)

func f(position [2]float64) [2]float64 {
	return [2]float64{0.5, 0.5}
}

func g(position [2]float64) [2]float64 {
	return [2]float64{0.5, 0.5}
}

func position_to_string(position [2]float64) string {
	return strconv.FormatFloat(position[0], 'g', -1, 64) + " " + strconv.FormatFloat(position[0], 'g', -1, 64) + "\n"
}

func ReadLine(r io.Reader, lineNum int) (line string, lastLine int, err error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		if lastLine == lineNum {
			return sc.Text(), lastLine, sc.Err()
		}
		lastLine++
	}
	return line, lastLine, io.EOF
}

// Run the simulation
func run_simulation(scheme schemes.NumericScheme,
	fp filepool.FilePool,
	path string,
	identifier string,
	nSteps int,
	stepSize float64) {
	f := fp.OpenFile(path + identifier)

	f.File.WriteString(position_to_string(scheme.GetPosition()))
	for i := 0; i < nSteps; i++ {
		scheme.Update(stepSize)
		f.File.WriteString(position_to_string(scheme.GetPosition()))
	}

	f.Close()
}

// Create a path, and anny specific subdirectories
func create_paths(path string, specific_paths []string) {
	for _, sp := range specific_paths {
		path := path + sp
		err := os.MkdirAll(path, 0775)
		check_error(err)
	}
}

// Check the errors and if there are any throw a panic
func check_error(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// First figure out the starting time
	t, err := time.Now().MarshalText()
	check_error(err)

	// Control Variables
	// Create storage
	path := "/dat/simulations/stochastic-particle-model/" + string(t) + "/"
	specific_paths := [3]string{"particles/", "steps/", "matrixplots/"}
	const filelimit = 1000

	// Create particles
	const nParticles = 10
	const nSteps = 2000
	const stepSize = float64(0.001)
	position := [2]float64{0.5, 0.5}

	// Create the scheme
	scheme := schemes.NewForwardEuler2D(0,
		position,
		f,
		g,
	)

	// Other stuff that needs to happen
	var wg sync.WaitGroup
	syscall.Umask(0)                      // We don't want the umask trumping our efforts.
	create_paths(path, specific_paths[:]) // Create all specific folders

	fp := filepool.NewFilePool(filelimit)

	// Step 1: Generate the data of each of the particles, and store that data in particlefiles
	for i := 0; i < nParticles; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			run_simulation(scheme,
				fp,
				path+specific_paths[0],
				"particle"+strconv.Itoa(i),
				nSteps,
				stepSize,
			)

		}()
	}

	// Wait for this step to finish before moving on
	wg.Wait()
	chFinishedSteps := make(chan int, 10)

	go func() {
		for finishedStep := range chFinishedSteps {
			fmt.Println("Compilation of step ", finishedStep, " is finished.")
		}
	}()

	// Step 2: Organise all data in steps.
	var swg sync.WaitGroup
	for s := 0; s <= nSteps; s++ {
		s := s
		wg.Add(1)

		c := make(chan string, nParticles) // Make a channel that will store all of the strings

		for p := 0; p < nParticles; p++ {
			swg.Add(1)
			p := p

			go func() {
				defer swg.Done()

				// Open the file if possible
				f := fp.OpenFile(path + specific_paths[0] + "particle" + strconv.Itoa(p))
				defer f.Close()
				// Read the step we need
				line, _, err := ReadLine(f.File, s)
				check_error(err)

				// Return the line
				c <- line + "\n"
			}()
		}

		swg.Wait() // wWit to finish all jobs before closing the channel
		close(c)   // Close the channel

		go func() {
			defer wg.Done()
			// Open the step-file
			f := fp.OpenFile(path + specific_paths[1] + "step" + strconv.Itoa(s))

			// Write in the file
			for p := range c {
				f.File.WriteString(p)
			}

			f.Close()

			chFinishedSteps <- s
		}()
	}

	wg.Wait() //Wait until all steps are compiled and saved.

}
