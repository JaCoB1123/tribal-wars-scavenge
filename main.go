package main

import (
	"flag"
	"fmt"
	"math"
	"time"
)

const baseTime = 23 * 60

var carry int
var maxScavenge int
var verbose bool
var debug bool

type raubzug struct {
	timeFact  float64
	carryFact float64
	name      string
}

func main() {
	flag.IntVar(&carry, "t", 1000, "Die Gesamte Tragekapazität der freien Einheiten, die aufgeteilt werden soll")
	flag.IntVar(&maxScavenge, "m", 4, "Falls noch nicht alle Raubzüge freigeschalten wurden, kann hier eine Zahl von 1 bis 4 übergeben werden")
	flag.BoolVar(&verbose, "verbose", false, "Mehr Informationen ausgeben")
	flag.BoolVar(&debug, "debug", false, "Technische Informationen ausgeben")
	flag.Parse()

	if maxScavenge > 4 || maxScavenge < 1 {
		fmt.Println("Der maximale Raubzug muss zwischen 1 und 4 liegen")
		return
	}

	step := 25
	for carry >= step*1000 {
		step = step * 10
	}
	for carry >= step*200 {
		step = step * 2
	}

	if carry%step != 0 {
		oldCarry := carry
		carry -= carry%step + step
		fmt.Printf("Tragekapazität %d ist nicht durch %d teilbar, wurde auf %d erhöht\n", oldCarry, step, carry)
	}

	if verbose {
		fmt.Printf("Using stepsize of %d\n", step)
	}

	var scavenges []raubzug

	scavenges = append(scavenges, raubzug{
		name:      "1",
		timeFact:  1,
		carryFact: 0.1,
	})
	scavenges = append(scavenges, raubzug{
		name:      "2",
		timeFact:  2.5,
		carryFact: 0.25,
	})
	scavenges = append(scavenges, raubzug{
		name:      "3",
		timeFact:  5,
		carryFact: 0.5,
	})
	scavenges = append(scavenges, raubzug{
		name:      "4",
		timeFact:  10,
		carryFact: 0.75,
	})

	ch := make(chan []int)

	total := carry
	var partial []int
	go func() {
		subset_sum(ch, total, partial, 0, step)
		close(ch)
	}()

	var score float64 = 0
	var bestThing []int
	for thing := range ch {
		var thisScore float64 = 0
		for i := 0; i < len(thing); i++ {
			_, _, sc := scavenges[i].calc(thing[i])
			thisScore += sc
		}

		if thisScore > score {
			if verbose {
				fmt.Printf("%7.0f res/h with config %v\n", thisScore, thing)
			}
			score = thisScore + 0.01
			bestThing = thing
		}
	}

	for i := 0; i < len(bestThing); i++ {
		seconds, carry, sc := scavenges[i].calc(bestThing[i])
		userTime := time.Duration(seconds) * time.Second
		fmt.Printf("Scavenge %s: %8d @ %7.0f res/h (%.0f resources in %s)\n", scavenges[i].name, bestThing[i], sc, carry, userTime.String())
	}
	fmt.Printf("Total                @ %7.0f res/h:\n", score)
}

func (scav raubzug) calc(total int) (float64, float64, float64) {
	time := math.Pow(float64(total)*scav.timeFact, 0.881) + baseTime

	carry := float64(total) * scav.carryFact
	score := carry / time * 60 * 60
	return time, carry, score
}

func subset_sum(result chan []int, max int, partial []int, partial_sum int, step int) {
	if debug {
		fmt.Printf("%v %v %v %v\n", max, partial, partial_sum, step)
	}
	if partial_sum == max {
		tmp := make([]int, len(partial))
		copy(tmp, partial)
		result <- tmp
	}
	if partial_sum >= max {
		return
	}
	if len(partial) == maxScavenge {
		return
	}
	for i := 0; i <= max; i += step {
		subset_sum(result, max, append(partial, i), partial_sum+i, int(math.Min(float64(step), float64(max-i))))
	}
}
