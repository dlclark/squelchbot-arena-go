package squelch

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"sync"

	"github.com/segmentio/ksuid"
)

type Tournament struct {
	matchCount      int
	gamesPerMatch   int
	playersPerMatch int
	targetScore     int
	entrants        []Player
}

func NewTournament(gpm, ppm, targetScore int, entrants []Player) *Tournament {
	return &Tournament{
		matchCount:      calcMatchCount(len(entrants), ppm),
		gamesPerMatch:   gpm,
		playersPerMatch: ppm,
		targetScore:     targetScore,
		entrants:        entrants,
	}
}

// GetMatchCount returns the number of matches that need to be played total for
// every player to play every other player an even number of times.
func (t Tournament) GetMatchCount() int {
	return t.matchCount
}

func (t *Tournament) Run() (*Results, error) {
	ranks := struct {
		sync.Mutex
		r []Points
	}{r: make([]Points, len(t.entrants))}

	wg := sync.WaitGroup{}

	entrantNames := make([]string, len(t.entrants))
	for i, p := range t.entrants {
		ranks.r[i] = Points{EntrantIndex: i}
		info, err := p.Info()
		if err != nil {
			// we couldn't get the name to start the tournament...that's bad, halt
			return nil, fmt.Errorf("Error getting info for entrant %v: %v", i, err)
		}
		entrantNames[i] = info.Name
	}

	// run full round-robin tournament with the entrants based on the
	// number of players in each game
	comb(len(t.entrants), t.playersPerMatch, func(players []int) {
		wg.Add(1)

		go func(players []int) {
			defer wg.Done()
			// make a match from the set of players
			p := make([]Player, t.playersPerMatch)

			// randomize our incoming player order and make a map
			plMap := rand.Perm(len(players))
			plNames := make([]string, len(players))

			for i := 0; i < len(plMap); i++ {
				p[i] = t.entrants[players[plMap[i]]]
				plNames[i] = entrantNames[players[plMap[i]]]
			}

			m := &match{
				players:      p,
				targetScore:  t.targetScore,
				gamesInMatch: t.gamesPerMatch,
				matchID:      ksuid.New().String(),
			}

			log.Printf("Match %v Start: Players %v", m.matchID, plNames)
			wins, _ := m.run()
			log.Printf("Match %v End: Wins %v", m.matchID, wins)
			// player with the most wins gets the match
			winner, highScore := 0, 0
			for i := 0; i < len(wins); i++ {
				// lookup our match index into the comb index into the entrant index
				entrantIdx := players[plMap[i]]
				// sum up wins
				ranks.Lock()
				ranks.r[entrantIdx].TotalWins += wins[i]
				ranks.r[entrantIdx].Matches++
				ranks.Unlock()

				// find highest score for points
				if wins[i] > highScore {
					winner, highScore = entrantIdx, wins[i]
				} else if wins[i] == highScore {
					// a tie -- nobody gets points
					winner = -1
				}
			}

			//no points for ties
			if winner > -1 {
				// 1 point for the winner, 0 for losers
				ranks.Lock()
				ranks.r[winner].Points++
				ranks.Unlock()
			}
		}(append([]int(nil), players...))
		// we need to pass in a clone of players
	})

	// wait for all matches to end
	wg.Wait()

	// sort the points and return
	sort.Slice(ranks.r, func(i, j int) bool {
		// backwards "less" so we sort high to low
		if ranks.r[i].Points == ranks.r[j].Points {
			return ranks.r[j].TotalWins < ranks.r[i].TotalWins
		}

		return ranks.r[j].Points < ranks.r[i].Points
	})

	return &Results{Points: ranks.r, EntrantNames: entrantNames}, nil
}

// emit all combinations of size m from set [0..n)
func comb(n, m int, emit func([]int)) {
	s := make([]int, m)
	last := m - 1
	var rc func(int, int)
	rc = func(i, next int) {
		for j := next; j < n; j++ {
			s[i] = j
			if i == last {
				emit(s)
			} else {
				rc(i+1, j+1)
			}
		}
		return
	}
	rc(0, 0)
}

type match struct {
	players             []Player
	playerNames         []string
	targetScore         int
	gamesInMatch        int
	matchID             string
	nextGameNumber      int
	startingPlayerIndex int
}

// Run is going to do blah
func (m *match) run() (wins, errs []int) {
	//setup results
	pc := len(m.players)
	wins = make([]int, pc)
	errs = make([]int, pc)

	// notify all players match begin
	for i, p := range m.players {
		err := p.MatchStart(m.matchID, 6, m.targetScore, m.gamesInMatch, i, m.playerNames)
		if err != nil {
			debug("Error on match start: %v\n", err)
			errs[i]++
		}
	}

	for i := 0; i < m.gamesInMatch; i++ {
		// make a new game
		m.nextGameNumber++
		m.startingPlayerIndex++
		if m.startingPlayerIndex >= len(m.players) {
			m.startingPlayerIndex = 0
		}
		g := NewGame(m.players, m.targetScore, m.matchID, strconv.Itoa(m.nextGameNumber), m.startingPlayerIndex)

		debug("Game %v/%v: Start", m.matchID, g.gameID)
		res, err := g.Run()
		if err != nil {
			debug("Game %v/%v: End with Error by player %v:%v", m.matchID, g.gameID, res.ErrIndex, err)
			errs[res.ErrIndex]++
		} else {
			debug("Game %v/%v: End with Win by player %v", m.matchID, g.gameID, res.WinnerIndex)
			wins[res.WinnerIndex]++
		}
	}

	// notify players match end
	for _, p := range m.players {
		p.MatchEnd(m.matchID, wins)
	}

	return
}

func calcMatchCount(playersTotal, ppm int) int {
	// https://en.wikipedia.org/wiki/Binomial_coefficient
	// calculates n choose k. Overflows are not detected
	n := playersTotal
	k := ppm

	if k > n {
		panic("Choose: k > n")
	}
	if k < 0 {
		panic("Choose: k < 0")
	}
	if n <= 1 || k == 0 || n == k {
		return 1
	}
	if newK := n - k; newK < k {
		k = newK
	}
	if k == 1 {
		return n
	}
	// Our return value, and this allows us to skip the first iteration.
	ret := n - k + 1
	for i, j := ret+1, 2; j <= k; i, j = i+1, j+1 {
		ret = ret * i / j
	}
	return ret
}

type Results struct {
	Points       []Points
	EntrantNames []string
}
type Points struct {
	EntrantIndex int
	Points       int
	Matches      int
	TotalWins    int
}

func (p Points) String() string {
	return fmt.Sprintf("{ Player: %v, Points: %v, Wins: %v }", p.EntrantIndex, p.Points, p.TotalWins)
}

type Match struct {
	gameCount   int
	targetScore int
	players     []Player
}
