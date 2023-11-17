package recurring

import (
	"fmt"
	"slices"
	"sort"
	"testing"

	"github.com/adrg/strutil/metrics"
	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	debug := false
	individualFailures := false
	data := GetFixtures(t, "monetr_sample_data_1.json")

	type SubResult struct {
		Dataset          string
		OverallAccuracy  float64
		CorrectMatches   int
		IncorrectMatches int
		TotalComparisons int
		LowestCorrect    float64
		HighestIncorrect float64
	}

	type MatrixResult struct {
		OverallAccuracy  float64
		CorrectMatches   int
		IncorrectMatches int
		TotalComparisons int
		DataSets         map[string]SubResult
		Comparator       string
	}

	allResults := map[string]MatrixResult{}

	comparors := map[string]TransactionNameComparator{
		"Hamming": &transactionComparatorBase{
			impl: &metrics.Hamming{
				CaseSensitive: false,
			},
		},
		"Levenshtein insert=1 replace=2 delete=1": &transactionComparatorBase{
			impl: &metrics.Levenshtein{
				CaseSensitive: false,
				InsertCost:    1,
				ReplaceCost:   2,
				DeleteCost:    1,
			},
		},
		"Jaro": &transactionComparatorBase{
			impl: &metrics.Jaro{
				CaseSensitive: false,
			},
		},
		"JaroWinkler": &transactionComparatorBase{
			impl: &metrics.JaroWinkler{
				CaseSensitive: false,
			},
		},
		"SmithWatermanGotoh gap=-0.1 match=1 mismatch=-0.5": &transactionComparatorBase{
			impl: &metrics.SmithWatermanGotoh{
				CaseSensitive: false,
				GapPenalty:    -0.1,
				Substitution: metrics.MatchMismatch{
					Match:    1,
					Mismatch: -0.5,
				},
			},
		},
		"SmithWatermanGotoh gap=-0.1 match=1 mismatch=-0.25": &transactionComparatorBase{
			impl: &metrics.SmithWatermanGotoh{
				CaseSensitive: false,
				GapPenalty:    -0.1,
				Substitution: metrics.MatchMismatch{
					Match:    1,
					Mismatch: -0.25,
				},
			},
		},
		"SmithWatermanGotoh gap=-0.25 match=1 mismatch=-0.25": &transactionComparatorBase{
			impl: &metrics.SmithWatermanGotoh{
				CaseSensitive: false,
				GapPenalty:    -0.25,
				Substitution: metrics.MatchMismatch{
					Match:    1,
					Mismatch: -0.25,
				},
			},
		},
		"SorensenDice ngram=0": &transactionComparatorBase{
			impl: &metrics.SorensenDice{
				CaseSensitive: false,
				NgramSize:     0,
			},
		},
		"SorensenDice ngram=1": &transactionComparatorBase{
			impl: &metrics.SorensenDice{
				CaseSensitive: false,
				NgramSize:     1,
			},
		},
		"SorensenDice ngram=2": &transactionComparatorBase{
			impl: &metrics.SorensenDice{
				CaseSensitive: false,
				NgramSize:     2,
			},
		},
		"SorensenDice ngram=3": &transactionComparatorBase{
			impl: &metrics.SorensenDice{
				CaseSensitive: false,
				NgramSize:     3,
			},
		},
		"SorensenDice ngram=4": &transactionComparatorBase{
			impl: &metrics.SorensenDice{
				CaseSensitive: false,
				NgramSize:     4,
			},
		},
		"Jaccard ngram=0": &transactionComparatorBase{
			impl: &metrics.Jaccard{
				CaseSensitive: false,
				NgramSize:     0,
			},
		},
		"Jaccard ngram=1": &transactionComparatorBase{
			impl: &metrics.Jaccard{
				CaseSensitive: false,
				NgramSize:     1,
			},
		},
		"Jaccard ngram=2": &transactionComparatorBase{
			impl: &metrics.Jaccard{
				CaseSensitive: false,
				NgramSize:     2,
			},
		},
		"Jaccard ngram=3": &transactionComparatorBase{
			impl: &metrics.Jaccard{
				CaseSensitive: false,
				NgramSize:     3,
			},
		},
		"Jaccard ngram=4": &transactionComparatorBase{
			impl: &metrics.Jaccard{
				CaseSensitive: false,
				NgramSize:     4,
			},
		},
		"OverlapCoefficient ngram=0": &transactionComparatorBase{
			impl: &metrics.OverlapCoefficient{
				CaseSensitive: false,
				NgramSize:     0,
			},
		},
		"OverlapCoefficient ngram=1": &transactionComparatorBase{
			impl: &metrics.OverlapCoefficient{
				CaseSensitive: false,
				NgramSize:     1,
			},
		},
		"OverlapCoefficient ngram=2": &transactionComparatorBase{
			impl: &metrics.OverlapCoefficient{
				CaseSensitive: false,
				NgramSize:     2,
			},
		},
		"OverlapCoefficient ngram=3": &transactionComparatorBase{
			impl: &metrics.OverlapCoefficient{
				CaseSensitive: false,
				NgramSize:     3,
			},
		},
	}

	type KnownGood struct {
		Name       string
		BaselineId uint64
		Matches    []uint64
	}

	testInput := []KnownGood{
		{
			// When I Work direct deposit items.
			Name:       "When I Work Direct Deposit",
			BaselineId: 290,
			Matches: []uint64{
				285,
				279,
				275,
				269,
				262,
				256,
				251,
				246,
				244,
				241,
				232,
				227,
				220,
				215,
				209,
				207,
				199,
				193,
				186,
				170,
			},
		},
		{
			Name:       "FreshBooks",
			BaselineId: 291,
			Matches: []uint64{
				280,
				270,
				257,
				245,
				234,
				223,
				211,
				200,
				188,
				173,
				161,
				149,
				137,
				126,
				114,
				104,
				81,
				71,
				60,
				50,
				39,
				27,
				14,
				293,
				608,
				747,
				768,
				794,
			},
		},
		{
			Name:       "Sentry",
			BaselineId: 286,
			Matches: []uint64{
				276,
				264,
				253,
				240,
				230,
				218,
				205,
				196,
				182,
				169,
				157,
				145,
				134,
				122,
				111,
				101,
				78,
				68,
				57,
				46,
				34,
				22,
				8,
				304,
				680,
				759,
				782,
				286,
			},
		},
		{
			Name:       "Google Cloud",
			BaselineId: 289,
			Matches: []uint64{
				284,
				283,
				278,
				274,
				273,
				268,
				261,
				260,
				255,
				250,
				249,
				243,
				238,
				237,
				233,
				228,
				226,
				221,
				216,
				214,
				210,
				206,
				203,
				198,
				194,
				192,
				187,
				180,
				178,
				176,
				177,
				164,
				165,
				166,
				153,
				151,
				152,
				139,
				140,
				141,
				130,
				129,
				116,
				118,
				117,
				105,
				106,
				95,
				96,
				// Goofy
				// 72,
				// 73,
				// 61,
				// 62,
				// 52,
				// 51,
				// 41,
				// 40,
				// 29,
				// 28,
				// 17,
				// 16,
				// 1,
				// 2,
				// 602,
				// 588,
				// 745,
				// 744,
				// 766,
				// Not goofy
				790,
			},
		},
		{
			Name:       "Goofy Google Cloud",
			BaselineId: 72,
			Matches: []uint64{
				73,
				61,
				62,
				52,
				51,
				41,
				40,
				29,
				28,
				17,
				16,
				1,
				2,
				602,
				588,
				745,
				744,
				766,

				// Deposit
				82,
			},
		},
		{
			Name:       "Treasury Prime Deposit",
			BaselineId: 185,
			Matches: []uint64{
				179,
				172,
				167,
				160,
				155,
				148,
				143,
				136,
				132,
				124,
				120,
				113,
				109,
				103,
				99,
				80,
				76,
				70,
				65,
				59,
				55,
				49,
				44,
				38,
				32,
				26,
				20,
				12,
				5,
				297,
				305,
				633,
				681,
				749,
				760,
				771,
				783,
				797,
			},
		},
	}
	getTransaction := func(id uint64) Transaction {
		for i := range data {
			txn := data[i]
			if txn.TransactionId == id {
				return txn
			}
		}
		panic("failed to find transaction with the specified ID")
	}

	type Score struct {
		A, B  Transaction
		Kind  string
		Score float64
	}

	desiredMatch := 0.5

	for name, compare := range comparors {
		for _, input := range testInput {
			t.Run(fmt.Sprintf("%s - %s", input.Name, name), func(t *testing.T) {
				subResult := SubResult{
					Dataset:          input.Name,
					OverallAccuracy:  0,
					CorrectMatches:   0,
					IncorrectMatches: 0,
					TotalComparisons: len(data),
				}
				baseline := getTransaction(input.BaselineId)
				var highestBad, lowestGood Score
				var correctMatches int
				for _, other := range data {
					if other.TransactionId == baseline.TransactionId {
						assert.EqualValues(t, 1, compare.CompareTransactionName(baseline, other), "comparing the same transaction should equal 1")
						correctMatches++
						continue
					}

					score := compare.CompareTransactionName(baseline, other)
					shouldMatch := slices.Contains(input.Matches, other.TransactionId)
					if shouldMatch {
						if individualFailures {
							if !assert.Greater(t, score, desiredMatch, "SHOULD MATCH! similar transactions should be at least 50% similar") {
								fmt.Printf("        	Kind: %s\n", name)
								fmt.Printf("        	Baseline (%d): %s\n", baseline.TransactionId, baseline.OriginalName)
								fmt.Printf("        	Other    (%d): %s\n", other.TransactionId, other.OriginalName)
							} else {
								correctMatches++
							}
						} else if score > desiredMatch {
							correctMatches++
						}
						if score < lowestGood.Score || lowestGood.Score == 0 {
							lowestGood = Score{
								A:     baseline,
								B:     other,
								Kind:  name,
								Score: score,
							}
						}
					} else {
						if individualFailures {
							if !assert.Less(t, score, desiredMatch, "SHOULD NOT MATCH! non similar transactions should be at most 50% similar") {
								fmt.Printf("        	Kind: %s\n", name)
								fmt.Printf("        	Baseline (%d): %s\n", baseline.TransactionId, baseline.OriginalName)
								fmt.Printf("        	Other    (%d): %s\n", other.TransactionId, other.OriginalName)
							} else {
								correctMatches++
							}
						} else if score < desiredMatch {
							correctMatches++
						}

						if score > highestBad.Score || highestBad.Score == 0 {
							highestBad = Score{
								A:     baseline,
								B:     other,
								Kind:  name,
								Score: score,
							}
						}
					}

					if debug {
						fmt.Printf("\n\tKind: %s\n", name)
						fmt.Printf("\tScore: %f\n", score)
						fmt.Printf("\tBaseline (%d): %s\n", baseline.TransactionId, baseline.OriginalName)
						fmt.Printf("\tOther    (%d): %s\n", other.TransactionId, other.OriginalName)
					}

				}

				subResult.CorrectMatches = correctMatches
				subResult.IncorrectMatches = len(data) - correctMatches
				subResult.OverallAccuracy = (float64(correctMatches) / float64(len(data))) * 100
				subResult.LowestCorrect = lowestGood.Score
				subResult.HighestIncorrect = highestBad.Score

				result, _ := allResults[name]
				result.Comparator = name
				result.IncorrectMatches += subResult.IncorrectMatches
				result.CorrectMatches += subResult.CorrectMatches
				result.TotalComparisons += subResult.TotalComparisons
				result.IncorrectMatches += subResult.IncorrectMatches
				if result.DataSets == nil {
					result.DataSets = map[string]SubResult{}
				}
				result.DataSets[input.Name] = subResult
				allResults[name] = result

				if individualFailures {
					assert.Greater(t, lowestGood.Score, highestBad.Score, "The lowest correct score must be higher than the highest incorrect score!")
				}

				fmt.Printf("\n=====================Comparison Results!=====================\n")
				fmt.Printf("\tTest: %s\n", t.Name())
				fmt.Printf("\tAccuracy: %d%s Correct: %d Total: %d\n\n", int((float64(correctMatches)/float64(len(data)))*100), "%", correctMatches, len(data))
				fmt.Printf("\tHighest Wrong Score!\n")
				fmt.Printf("\tKind: %s\n", highestBad.Kind)
				fmt.Printf("\tScore: %f\n", highestBad.Score)
				fmt.Printf("\tBaseline (%d): %s\n", highestBad.A.TransactionId, highestBad.A.OriginalName)
				fmt.Printf("\tOther    (%d): %s\n", highestBad.B.TransactionId, highestBad.B.OriginalName)
				fmt.Printf("\n\tLowest Correct Score!\n")
				fmt.Printf("\tKind: %s\n", lowestGood.Kind)
				fmt.Printf("\tScore: %f\n", lowestGood.Score)
				fmt.Printf("\tBaseline (%d): %s\n", lowestGood.A.TransactionId, lowestGood.A.OriginalName)
				fmt.Printf("\tOther    (%d): %s\n\n", lowestGood.B.TransactionId, lowestGood.B.OriginalName)
			})
		}
	}

	finalResults := make([]MatrixResult, 0, len(allResults))
	for key := range allResults {
		result := allResults[key]
		result.OverallAccuracy = (float64(result.CorrectMatches) / float64(result.TotalComparisons)) * 100
		finalResults = append(finalResults, result)
	}
	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].OverallAccuracy > finalResults[j].OverallAccuracy
	})
	for i := range finalResults {
		result := finalResults[i]

		fmt.Printf("\n=========================================================================================\n")
		fmt.Printf("Comparator: %s\n", result.Comparator)
		fmt.Printf("Accuracy: %f%s\n", result.OverallAccuracy, "%")
		falsePositivesObserved := false
		for _, sub := range result.DataSets {
			if sub.HighestIncorrect > sub.LowestCorrect {
				falsePositivesObserved = true
				break
			}
		}
		if falsePositivesObserved {
			fmt.Printf("False Positives Observed: YES!\n")
		} else {
			fmt.Printf("False Positives Observed: NO!\n")
		}
	}
}
