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

	"github.com/AlexdeRuijter/stochastic-particle-model/filepool"
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

func dg(position [2]float64) [2]float64 {
	var r [2]float64

	for i, v := range position {
		r[i] = -math.Pi * math.Sin(math.Pi*v) / math.Sqrt(8+8*math.Cos(math.Pi*v))
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

type Position struct {
	x float64
	y float64
	t float64
}

func main() {
	var wg sync.WaitGroup

	// First figure out the starting time
	t, err := time.Now().MarshalText()
	check_error(err)

	// Control Variables
	// Create storage
	path := "/dat/simulations/stochastic-particle-model/" + string(t) + "/"
	specific_paths := [4]string{"particles/", "steps/", "matrixplots/", "multiplot"}
	const filelimit = 1000

	// How many steps
	syscall.Umask(0)                      // We don't want the umask trumping our efforts.
	create_paths(path, specific_paths[:]) // Create all specific folders
	fp := filepool.NewFilePool(filelimit) // Create the filepool

	a := make([]string, 0, 198)
	for i := 10; i <= 1000; i = i + 5 {
		a = append(a, "plot"+strconv.Itoa(i)+"/")

	}

	create_paths(path, a)

	for j := 0; j <= 198; j++ {
		wg.Add(1)

		i := j*5 + 10

		go generate_paths(i, fp, path, a[j], &wg)
	}

	wg.Wait()

}

func generate_paths(nSteps int, fp filepool.FilePool, path string, specific_paths string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create particles
	const nParticles = 500
	var stepSize = 0.1 / float64(nSteps)

	var position = [2]float64{0.5, 0.5}

	C := make([][]Position, nParticles)
	for i := 0; i < nParticles; i++ {
		C = append(C, make([]Position, 0, nSteps))
	}

	for i := 0; i < nParticles; i++ {
		i := i
		{
			// Create the scheme
			scheme := schemes.NewForwardEuler2D(0,
				position,
				f,
				g,
			)

			f := fp.OpenFile(path + specific_paths + "particle" + strconv.Itoa(i))
			f.File.WriteString(strconv.FormatFloat(0., 'f', -1, 64) + " " + position_to_string(scheme.GetPosition()))
			for j := 0; j < nSteps; j++ {
				scheme.Update(stepSize)
				pos := scheme.GetPosition()
				t := scheme.GetRandomState().GetTime()

				// Save the run
				f.File.WriteString(strconv.FormatFloat(t, 'f', -1, 64) + " " + position_to_string(pos))

				// Save all runs
				p := Position{pos[0], pos[1], t}
				C[i] = append(C[i], p)

			}

			f.Close()
		}
	}

	//fmt.Println(C[1][1].x, C[1][1].x)
}
