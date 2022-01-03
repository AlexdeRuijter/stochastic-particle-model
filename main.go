package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/AlexdeRuijter/stochastic-particle-model/analysis"
	"github.com/AlexdeRuijter/stochastic-particle-model/filepool"
	"github.com/AlexdeRuijter/stochastic-particle-model/matrixplots"
	"github.com/AlexdeRuijter/stochastic-particle-model/schemes"
)

func f(position [2]float64) [2]float64 {
	return [2]float64{0.0, 0.0}
}

func g(position [2]float64) [2]float64 {
	return [2]float64{0.5, 0.5}
}

func position_to_string(position [2]float64) string {
	return strconv.FormatFloat(position[0], 'g', -1, 64) + " " + strconv.FormatFloat(position[0], 'g', -1, 64) + "\n"
}

func string_to_position(s string) [2]float64 {
	var r [2]float64
	subs := strings.Split(s, " ")
	for i, s := range subs {
		f, err := strconv.ParseFloat(s, 64)
		check_error(err)
		r[i] = f
	}
	return r
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
	const nParticles = 100
	const nSteps = 1000
	const stepSize = float64(0.001)
	position := [2]float64{0.5, 0.5}

	// The respective scheme is created in step 1!

	// Other stuff that needs to happen
	var wgStep1, wgStep2, wgStep3 sync.WaitGroup

	syscall.Umask(0)                      // We don't want the umask trumping our efforts.
	create_paths(path, specific_paths[:]) // Create all specific folders

	fp := filepool.NewFilePool(filelimit)

	// Step 1: Generate the data of each of the particles, and store that data in particlefiles
	for i := 0; i < nParticles; i++ {
		i := i
		wgStep1.Add(1)
		go func() {
			defer wgStep1.Done()
			// Create the scheme
			scheme := schemes.NewForwardEuler2D(0,
				position,
				f,
				g,
			)

			run_simulation(scheme,
				fp,
				path+specific_paths[0],
				"particle"+strconv.Itoa(i),
				nSteps,
				stepSize,
			)

		}()
	}

	// Wait for step 1 to finish before moving on
	wgStep1.Wait()

	// Prepare setting up step 3, so it runs as soon as step 2 is finished
	chFinishedSteps := make(chan int, 10)

	//Launch the independent step 3
	wgStep3.Add(1)
	go func() {
		defer wgStep3.Done()
		for finishedStep := range chFinishedSteps {
			wgStep3.Add(1)
			step := finishedStep
			go func() {
				defer wgStep3.Done()
				var x, y [nParticles]float64

				// Open the datafile
				f := fp.OpenFile(path + specific_paths[1] + "step" + strconv.Itoa(step))
				scanner := bufio.NewScanner(f.File)
				lineNumber := 0

				// Create the datafiles x and y
				for scanner.Scan() {
					pos := string_to_position(scanner.Text())
					x[lineNumber], y[lineNumber] = pos[0], pos[1]
					lineNumber++
				}

				xMV := analysis.CalculateMeanAndVariation(x[:])
				yMV := analysis.CalculateMeanAndVariation(y[:])

				// Printing instead of saving for now:
				fmt.Println("x: ", x, " y: ", y, " xMV: ", xMV, " yMV: ", yMV)
				M := matrixplots.Histogram2D(x[:], y[:], 10, -1, 1, -1, 1)
				fmt.Println("Matrix form:")
				for _, arr := range M {
					fmt.Println(arr)
				}
				fmt.Println()

			}()
		}
	}()

	// Step 2: Organise all data in steps.
	for s := 0; s <= nSteps; s++ {
		s := s
		wgStep1.Add(1)

		c := make(chan string, nParticles) // Make a channel that will store all of the strings

		for p := 0; p < nParticles; p++ {
			wgStep2.Add(1)
			p := p

			go func() {
				defer wgStep2.Done()

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

		wgStep2.Wait() // wWit to finish all jobs before closing the channel
		close(c)       // Close the channel

		go func() {
			defer wgStep1.Done()
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

	wgStep1.Wait() //Wait until all steps are compiled and saved.
	wgStep2.Wait()
	close(chFinishedSteps)

	wgStep3.Wait()

}
