package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/AlexdeRuijter/stochastic-particle-model/analysis"
	"github.com/AlexdeRuijter/stochastic-particle-model/filepool"
	"github.com/AlexdeRuijter/stochastic-particle-model/schemes"
	"github.com/sbinet/go-gnuplot"
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
	return slice_to_stringf64(position[:]) + "\n"
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
	s = strings.TrimSpace(s)

	return s
}

func slice_to_stringi64(slice []int64) string {
	var s string

	for _, v := range slice {
		s += strconv.FormatInt(v, 10) + " "

	}
	s = strings.TrimSpace(s)

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
	specific_paths := [4]string{"particles/", "steps/", "matrixplots/", "multiplot/"}
	const filelimit = 1000

	// How many steps
	syscall.Umask(0)                      // We don't want the umask trumping our efforts.
	create_paths(path, specific_paths[:]) // Create all specific folders
	fp := filepool.NewFilePool(filelimit) // Create the filepool

	// a := make([]string, 0, 199)
	// for i := 5; i <= 1000; i = i + 5 {
	// 	a = append(a, "plot"+strconv.Itoa(i)+"/")

	// }

	// create_paths(path, a)

	// for j := 0; j <= 198; j++ {
	// 	wg.Add(1)

	// 	i := j*5 + 5

	// 	go generate_paths(i, fp, path, a[j], &wg)
	// }

	fp.Wait()

	var position = [2]float64{0.5, 0.5}
	X := make([]float64, 1000)
	Y := make([]float64, 1000)
	T := make([]float64, 1000)

	for i := 1; i < 1000; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()

			dt := 1. / float64(i)

			for j := 0; j < 100; j++ {
				scheme := schemes.NewForwardEuler2D(1,
					position,
					f,
					g,
				)
				xflag := false
				yflag := false

				for k := 0; k < i*10; k++ {

					if !xflag || !yflag {
						scheme.Update(dt)
					}
					pos := scheme.GetPosition()

					if !xflag && math.Abs(pos[0]) > 1. {
						xflag = true
					}
					if !yflag && math.Abs(pos[0]) > 1. {
						yflag = true
					}

				}
				if xflag {
					X[i-1] += 1
				}
				if yflag {
					Y[i-1] += 1
				}
			}

			T[i-1] = dt
		}()

	}

	wg.Wait()

	// Create a plot
	fname := ""
	persist := true
	debug := true

	p, err := gnuplot.NewPlotter(fname, persist, debug)
	if err != nil {
		err_string := fmt.Sprintf("** err: %v\n", err)
		panic(err_string)
	}
	defer p.Close()

	p.CheckedCmd("set title 'Numerical stability Forward Euler Scheme'")

	p.CheckedCmd(`set xlabel 'dt'`)
	//p.CheckedCmd("set log x")
	//p.CheckedCmd("set log y")
	p.CheckedCmd("set key left top")

	p.SetStyle("lines")

	p.PlotXY(T, X, `Left x-Domain`)
	p.PlotXY(T, Y, `Left y-Domain`)

	p.CheckedCmd("set output")
}

func generate_paths(nSteps int, fp filepool.FilePool, path string, specific_paths string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create particles
	const nParticles = 500
	var stepSize = .5 / float64(nSteps)

	var position = [2]float64{0.5, 0.5}

	C := make([][]Position, nParticles)
	for i := 0; i < nParticles; i++ {
		C = append(C, make([]Position, 0, nSteps))
	}

	for i := 0; i < nParticles; i++ {
		i := i
		{
			// Create the scheme
			scheme := schemes.NewMilstein(0,
				position,
				f,
				g,
				dg,
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

	f := fp.OpenFile(path + "multiplot/" + "analysis" + strconv.Itoa(nSteps))

	for i := 0; i < nSteps; i++ {
		X := make([]float64, 0, nParticles)
		Y := make([]float64, 0, nParticles)

		for j := 0; j < nParticles; j++ {
			p := C[j][i]

			X = append(X, p.x)
			Y = append(Y, p.y)
		}

		Xmv := analysis.CalculateMeanAndVariation(X[:])
		Ymv := analysis.CalculateMeanAndVariation(Y[:])

		f.File.WriteString(strconv.FormatFloat(C[0][i].t, 'f', -1, 64) + " " + slice_to_stringf64(Xmv[:]) + " " + slice_to_stringf64(Ymv[:]) + "\n")
	}

	f.Close()

	//fmt.Println(C[1][1].x, C[1][1].x)
}
