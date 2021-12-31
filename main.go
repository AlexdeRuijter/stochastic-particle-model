package main

import (
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/AlexdeRuijter/stochastic-particle-model/filepool"
	"github.com/AlexdeRuijter/stochastic-particle-model/schemes"
)

func f(position [2]float64) [2]float64 {
	return position
}

func g(position [2]float64) [2]float64 {
	return position
}

func position_to_string(position [2]float64) string {
	return strconv.FormatFloat(position[0], 'g', -1, 64) + " " + strconv.FormatFloat(position[0], 'g', -1, 64) + "\n"
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
	filelimit := uint64(1000)

	// Create particles
	nParticles := 100000
	nSteps := 100
	stepSize := float64(0.001)
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
	create_paths(path, specific_paths[:]) //

	fp := filepool.NewFilePool(filelimit)

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
	wg.Wait()

}
