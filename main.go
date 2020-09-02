package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/dlclark/squelchbot-arena-go/localbot"
	"github.com/dlclark/squelchbot-arena-go/squelch"
)

// our flags
var (
	//+urls
	gpm = flag.Int("gpm", 10, "games per match")
	ppm = flag.Int("ppm", -1, "players per match")
)

func main() {
	var us urls
	flag.Var(&us, "url", "a client bot URL. Repeatable.")
	flag.Parse()

	// set related defaults
	if *ppm == -1 {
		*ppm = len(us)
	}

	// validate our inputs
	if err := validateInputs(us, *gpm, *ppm); err != nil {
		log.Fatalf("Invalid input: %v", err)
	}

	// setup our match
	// start running games based on number of configured concurrent count
	p := make([]squelch.Player, len(us))

	/*for i, u := range us {
		p[i] = squelch.NewApiPlayer(u)
	}*/
	p = []squelch.Player{
		localbot.NewLocalBotPlayer("Local1"),
		localbot.NewLocalBotPlayer("Local2"),
	}

	rand.Seed(time.Now().UTC().UnixNano())

	t := squelch.NewTournament(*gpm, *ppm, 5000, p)

	// number of matches to have an even tournament
	mc := t.GetMatchCount()

	// total game count
	totalGc := mc * *gpm

	// print our summary line
	fmt.Printf("Starting tournament with %v entrants, %v players per match, %v matches totaling %v games.\n", len(p), *ppm, mc, totalGc)

	// start the tournament!
	r, err := t.Run()
	if err != nil {
		fmt.Printf("An error running the tournament: %v\n", err)
	}

	if r == nil {
		return
	}

	// output final points
	fmt.Println("Ranks:")
	for i, s := range r.Points {
		fmt.Printf("\t%v: %v (won %4.1f%% match, %4.1f%% game)\n", i+1, r.EntrantNames[s.EntrantIndex],
			float64(s.Points*100.0)/float64(s.Matches), float64(s.TotalWins*100.0)/float64(s.Matches**gpm))
	}
}

func validateInputs(us urls, gpm, ppm int) error {
	if len(us) < 2 {
		return errors.New("at least 2 URLs are required")
	}

	if gpm < 1 {
		return errors.New("at least 1 game per match is required")
	}

	if ppm < 2 {
		return errors.New("at least 2 players per match are required")
	}

	if ppm > len(us) {
		return fmt.Errorf("players per match cannot exceed provided URL count (%v)", len(us))
	}

	return nil
}
