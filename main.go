package main

import (
	"bufio"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/AlexdeRuijter/stochastic-particle-model/analysis"
	"github.com/AlexdeRuijter/stochastic-particle-model/bytes"
	"github.com/AlexdeRuijter/stochastic-particle-model/filepool"
	"github.com/AlexdeRuijter/stochastic-particle-model/matrixplots"
	"github.com/AlexdeRuijter/stochastic-particle-model/schemes"
)

func f(position [2]float64) [2]float64 {
	var r [2]float64
	x := position[0]
	y := position[1]

	r[0] = ((x*x-1)*y+5*math.Cos(math.Pi*x)+5)/(5*x+15) - math.Pi*math.Sin(math.Pi*x)
	r[1] = ((y*y-1)*x)/(5*x+15) - math.Pi*math.Sin(math.Pi*y)

	return r
}

func g(position [2]float64) [2]float64 {
	var r [2]float64

	for i, v := range position {
		r[i] = math.Sqrt(2 + 2*math.Cos(math.Pi*v))
	}

	return r
}

func position_to_string(position [2]float64) string {
	return slice_to_stringf64(position[:])
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

func slice_to_stringf64(slice []float64) string {
	var s string

	for _, v := range slice {
		s += strconv.FormatFloat(v, 'g', -1, 64) + " "

	}
	s = strings.TrimSpace(s) + "\n"

	return s
}

func slice_to_stringi64(slice []int64) string {
	var s string

	for _, v := range slice {
		s += strconv.FormatInt(v, 10) + " "

	}
	s = strings.TrimSpace(s) + "\n"

	return s
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
	stepSize float64,
	C []chan [2][8]byte) {

	f := fp.OpenFile(path + identifier)

	f.File.WriteString(position_to_string(scheme.GetPosition()))
	for i := 0; i < nSteps; i++ {
		var passThrough [2][8]byte
		scheme.Update(stepSize)
		pos := scheme.GetPosition()
		for j, f := range pos {
			passThrough[j] = bytes.Float64_to_bytes(f)
		}
		C[i] <- passThrough
		f.WriteBytes(passThrough[:])

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

type Step struct {
	Channel chan [2][8]byte
	Number  int
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
	const nParticles = 50000
	const nSteps = 1000
	const stepSize = float64(0.0001)
	position := [2]float64{0.5, 0.5}

	// The respective scheme is created in step 1!
	var C [nSteps]chan [2][8]byte
	for i := range C {
		C[i] = make(chan [2][8]byte, nParticles)
	}

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
				C[:],
			)
		}()
	}

	// Wait for step 1 to finish before moving on
	wgStep1.Wait()

	for i := 0; i < nSteps; i++ {
		close(C[i])
	}

	// Prepare setting up step 3, so it runs as soon as step 2 is finished
	chFinishedSteps := make(chan Step, nSteps)

	//Launch the independent step 3
	wgStep3.Add(1)
	go func() {
		defer wgStep3.Done()
		for finishedStep := range chFinishedSteps {
			wgStep3.Add(1)
			step := finishedStep
			ch := step.Channel
			go func() {
				defer wgStep3.Done()
				var x, y [nParticles]float64

				var i int
				for b := range ch {
					fSlice := bytes.Position_from_bytesarray(b)
					x[i] = fSlice[0]
					y[i] = fSlice[1]
					i++
				}

				xMV := analysis.CalculateMeanAndVariation(x[:])
				yMV := analysis.CalculateMeanAndVariation(y[:])

				/*
					// Printing xMV and yMV for debugging
						fmt.Println("x: ", x, " y: ", y, " xMV: ", xMV, " yMV: ", yMV)
				*/

				M := matrixplots.Histogram2D(x[:], y[:], 500, -1, 1, -1, 1)

				/*
					//For printing the Matrix:
					fmt.Println("Matrix form:")
					for _, arr := range M {
						fmt.Println(arr)
					}
					fmt.Println()
				*/

				f := fp.OpenFile(path + specific_paths[2] + "summary_step" + strconv.Itoa(step.Number))

				f.File.WriteString("# X m v " + position_to_string(xMV))
				f.File.WriteString("# Y m v " + position_to_string(yMV))

				for _, arr := range M {
					f.File.WriteString(slice_to_stringi64(arr))
				}

				f.Close()
			}()
		}
	}()

	// Step 2: Organise all data in steps.
	for s := 0; s < nSteps; s++ {
		s := s
		wgStep1.Add(1)

		c := C[s]

		go func() {
			defer wgStep1.Done()
			g := make(chan [2][8]byte, nParticles)
			bytes := make([][8]byte, 0, nParticles*2)
			// Open the step-file
			f := fp.OpenFile(path + specific_paths[1] + "step" + strconv.Itoa(s))

			// Write in the file
			for p := range c {
				g <- p
				for _, b := range p {
					bytes = append(bytes, b)
				}
			}
			close(g)

			f.WriteBytes(bytes)
			f.Close()

			step := Step{Channel: g, Number: s}

			chFinishedSteps <- step
		}()
	}

	wgStep1.Wait() //Wait until all steps are compiled and saved.
	wgStep2.Wait()
	close(chFinishedSteps)

	wgStep3.Wait()

}
