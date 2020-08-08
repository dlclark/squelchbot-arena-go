package squelch

import (
	"math"
	"math/big"
	"sort"
	"strings"
)

// ScoringOption is a single option for taking points from a roll
type ScoringOption struct {
	ID        string
	DieValues string
	Points    int
}

//var optionsSync = sync.Once{}
var allOptions map[string][]ScoringOption

func init() {
	// generate our dice scoring options once
	//optionsSync.Do(func() {
	allOptions = getAllOptions()
	//})
}

func getAllOptions() map[string][]ScoringOption {
	//a given serial dice value gives you a list of scoring options
	// 11 choose 6 + 10 choose 5 + 9 choose 4 + 8 choose 3 + 7 choose 2 + 6 choose 1
	list := make(map[string][]ScoringOption, 923)

	debug("Making options...")
	// iterate every 6-dice combo

	for s := range generateCombinations("123456", 6) {
		//sort the "roll", see if it's in our list already
		data := []rune(s)
		sort.Slice(data, func(i, j int) bool {
			return data[i] < data[j]
		})

		d := string(data)
		if _, ok := list[d]; ok {
			continue
		}

		// not in our list, then calculate all sub-rolls of this one and get the scores for them
		list[s] = score(d)
	}

	debug("Done")
	return list
}

// generates a unique list of dice combinations for the given input
// for example: given an ordered slice [1,1,2,3]
// 	return lists [1] [1,1] [1,2] [1,3] [1,1,2] [1,1,3] [1,2,3] [1,1,2,3] [2] [2,3] [3]
// 	skips dupe lists [1] [1,2] [1,3] [1,2,3]
// example: [1,2,2,3]
// 	return lists [1] [1.2] [1,3] [1,2,2] [1,2,3] [1,2,2,3] [2] [2,2] [2,3] [2,2,3] [3]
// example: [1,2]
//	returns lists [1] [1,2] [2]
func generateLists(s string) <-chan string {
	c := make(chan string)

	go func(c chan string) {
		defer close(c)
		cache := make(map[string]struct{})
		//don't send back blank
		cache[""] = struct{}{}

		data := new(big.Int)
		one := big.NewInt(1)

		max := int(math.Pow(2, float64(len(s))))

		for i := 1; i < max; i++ {
			data.Add(data, one)

			// all the combinations of 0 and 1 for each place in the string
			// 0's aren't included, 1's are in our final string, check the cache and return it
			var p string
			for i := 0; i < len(s); i++ {
				if data.Bit(i) == 1 {
					p += string(s[i])
				}
			}

			if _, ok := cache[p]; !ok {
				c <- p
				cache[p] = struct{}{}
			}
		}

	}(c)

	return c
}

// make combinations from a unique alphabet up to a specific length.
// example:given alphabet 123, length 2 will generate strings 1 2 3 1 13 23
func generateCombinations(alphabet string, length int) <-chan string {
	c := make(chan string)

	// Starting a separate goroutine that will create all the combinations,
	// feeding them to the channel c
	go func(c chan string) {
		defer close(c) // Once the iteration function is finished, we close the channel

		addLetter(c, "", ' ', alphabet, length) // We start by feeding it an empty string
	}(c)

	return c // Return the channel to the calling function
}

func addLetter(c chan string, combo string, prevChar rune, alphabet string, length int) {
	// Check if we reached the length limit
	// If so, we just return without adding anything
	if length <= 0 {
		return
	}

	var newCombo string
	for _, ch := range alphabet {
		// if the new char is equal or later than our ending combo then we can do it
		if combo != "" {
			if prevChar > ch {
				continue
			}
		}
		newCombo = combo + string(ch)
		c <- newCombo
		addLetter(c, newCombo, ch, alphabet, length-1)
	}
}

// give set of sorted dice, return the list of possible scores
// including subsets of dice
func score(dice string) []ScoringOption {
	// iterate the subsets of dice and score with scoreUseAll
	opts := []ScoringOption{}

	//get all possible subsets (powerset)
	for sub := range generateLists(dice) {
		opts = append(opts, scoreUseAll(sub)...)
	}

	return sortOptions(opts)
}

// Sorter to get dice options in the highest point and fewest dice order
func sortOptions(opts []ScoringOption) []ScoringOption {
	//sort highest points first
	data := scoringOptionSlice(opts)
	sort.Sort(sort.Reverse(data))

	return []ScoringOption(data)
}

type scoringOptionSlice []ScoringOption

func (p scoringOptionSlice) Len() int { return len(p) }
func (p scoringOptionSlice) Less(i, j int) bool {
	pi := p[i]
	pj := p[j]
	if pi.Points == pj.Points {
		// second sort by number of dice used
		return len(pj.DieValues) < len(pi.DieValues)
	}
	return pi.Points < pj.Points
}
func (p scoringOptionSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// give set of sorted dice, return the list of possible scores
// that consume ALL the dice
func scoreUseAll(dice string) []ScoringOption {
	//Rules:
	// 1s: 100
	// 5:  50
	// 3 of a kind: Value * 100 (1 is 1000)
	// Triple Double: 750
	// 123456: 1500 points

	if dice == "123456" {
		// only option here
		return []ScoringOption{{DieValues: dice, Points: 1500}}
	}

	// triple doubles and 123456 use all dice
	remainingDice := dice

	// check each of our numbers 1-6 for instance counts, if we get 3 pairs
	// then that's triple doubles
	pairCt := 0
	curOpt := ScoringOption{}

	for _, side := range "123456" {
		side := string(side)
		ct := strings.Count(remainingDice, side)

		if ct == 6 {
			// special case, two three of a kind and triple doubles
			return []ScoringOption{
				{DieValues: remainingDice, Points: threeOfAKindPoints(side) * 2},
				{DieValues: remainingDice, Points: 750},
			}
		}

		if ct == 4 {
			//we have 2 pair
			pairCt += 2
		}

		if ct >= 3 {
			// three of a kind
			curOpt.append(ScoringOption{DieValues: strings.Repeat(side, 3), Points: threeOfAKindPoints(side)})
			remainingDice = strings.Replace(remainingDice, side, "", 3)
		}

		if ct == 2 {
			//we have a pair
			pairCt++
		}

		// recount now that we've taken care of sets
		ct = strings.Count(remainingDice, side)

		if side == "5" {
			// add on 5s
			curOpt.append(ScoringOption{DieValues: strings.Repeat(side, ct), Points: 50 * ct})
			remainingDice = strings.Replace(remainingDice, side, "", ct)
		} else if side == "1" {
			// add on 1s
			curOpt.append(ScoringOption{DieValues: strings.Repeat(side, ct), Points: 100 * ct})
			remainingDice = strings.Replace(remainingDice, side, "", ct)
		}
	}

	// triple doubles!
	if pairCt == 3 {
		// if we have remaining dice then this is our only option
		if remainingDice != "" {
			return []ScoringOption{{DieValues: dice, Points: 750}}
		}
		// return other option with this one
		return []ScoringOption{curOpt, {DieValues: dice, Points: 750}}
	}

	if remainingDice == "" && curOpt.Points > 0 {
		return []ScoringOption{curOpt}
	}

	// no way to use all dice, no scoring for this exact set
	return []ScoringOption{}
}

func threeOfAKindPoints(side string) int {
	switch side {
	case "1":
		return 1000
	case "2":
		return 200
	case "3":
		return 300
	case "4":
		return 400
	case "5":
		return 500
	case "6":
		return 600
	}
	return -1
}

func (s *ScoringOption) append(n ScoringOption) {
	s.Points += n.Points
	s.DieValues += n.DieValues
}
