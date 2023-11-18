# Recurring

This package contains the code necessary for detecting recurring transactions.

## Compare

In order to compare two transactions to determine whether or not they are the same or similar a text comparison
algorithm is used. This algorithm is chosen based on tests to analyze its accuracy against a dataset with a variety
of different parameters. At the time of writing this the `JaroWinkler` text comparison algorithm without strings
adjusted to be equal lengths and a success threshold of `0.83` is the best performing configuration.

```text
Comparator:          JaroWinkler
Accuracy:            99.730094% Total Comparisons Made:   66690
Threshold:           0.830000
False Positives:     2 0.002999%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348
```

The top several configurations are listed below, from least accurate to most accurate:

```text
=========================================================================================
Comparator:          Jaro EqualLengths
Accuracy:            98.575499% Total Comparisons Made:   66690
Threshold:           0.730000
False Positives:     187 0.280402%
False Negatives:     763 1.144100%
Lowest Correct:      0.500185
Highest Incorrect:   0.786890

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            98.582996% Total Comparisons Made:   66690
Threshold:           0.720000
False Positives:     771 1.156095%
False Negatives:     174 0.260909%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          Levenshtein insert=1 replace=2 delete=1 EqualLengths
Accuracy:            98.605488% Total Comparisons Made:   66690
Threshold:           0.450000
False Positives:     241 0.361374%
False Negatives:     689 1.033138%
Lowest Correct:      -0.279070
Highest Incorrect:   0.578947

=========================================================================================
Comparator:          SmithWatermanGotoh gap=-0.1 match=1 mismatch=-0.25 EqualLengths
Accuracy:            98.608487% Total Comparisons Made:   66690
Threshold:           0.690000
False Positives:     241 0.361374%
False Negatives:     687 1.030139%
Lowest Correct:      0.286047
Highest Incorrect:   0.747368

=========================================================================================
Comparator:          SmithWatermanGotoh gap=-0.1 match=1 mismatch=-0.5 EqualLengths
Accuracy:            98.608487% Total Comparisons Made:   66690
Threshold:           0.690000
False Positives:     241 0.361374%
False Negatives:     687 1.030139%
Lowest Correct:      0.286047
Highest Incorrect:   0.747368

=========================================================================================
Comparator:          SmithWatermanGotoh gap=-0.25 match=1 mismatch=-0.25
Accuracy:            98.638477% Total Comparisons Made:   66690
Threshold:           0.700000
False Positives:     736 1.103614%
False Negatives:     172 0.257910%
Lowest Correct:      0.322727
Highest Incorrect:   0.886364

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            98.639976% Total Comparisons Made:   66690
Threshold:           0.840000
False Positives:     1 0.001499%
False Negatives:     906 1.358525%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            98.642975% Total Comparisons Made:   66690
Threshold:           0.590000
False Positives:     18 0.026991%
False Negatives:     887 1.330034%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

=========================================================================================
Comparator:          Levenshtein insert=1 replace=2 delete=1
Accuracy:            98.660969% Total Comparisons Made:   66690
Threshold:           0.590000
False Positives:     72 0.107962%
False Negatives:     821 1.231069%
Lowest Correct:      0.081395
Highest Incorrect:   0.648649

=========================================================================================
Comparator:          SorensenDice ngram=2
Accuracy:            98.666967% Total Comparisons Made:   66690
Threshold:           0.740000
False Positives:     18 0.026991%
False Negatives:     871 1.306043%
Lowest Correct:      0.316547
Highest Incorrect:   0.746269

=========================================================================================
Comparator:          SorensenDice ngram=2
Accuracy:            98.666967% Total Comparisons Made:   66690
Threshold:           0.730000
False Positives:     18 0.026991%
False Negatives:     871 1.306043%
Lowest Correct:      0.316547
Highest Incorrect:   0.746269

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            98.666967% Total Comparisons Made:   66690
Threshold:           0.580000
False Positives:     18 0.026991%
False Negatives:     871 1.306043%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

=========================================================================================
Comparator:          Jaro
Accuracy:            98.672964% Total Comparisons Made:   66690
Threshold:           0.710000
False Positives:     715 1.072125%
False Negatives:     170 0.254911%
Lowest Correct:      0.591939
Highest Incorrect:   0.827712

=========================================================================================
Comparator:          OverlapCoefficient ngram=4
Accuracy:            98.696956% Total Comparisons Made:   66690
Threshold:           0.650000
False Positives:     585 0.877193%
False Negatives:     284 0.425851%
Lowest Correct:      0.230769
Highest Incorrect:   0.793103

=========================================================================================
Comparator:          Jaro EqualLengths
Accuracy:            98.707452% Total Comparisons Made:   66690
Threshold:           0.720000
False Positives:     259 0.388364%
False Negatives:     603 0.904184%
Lowest Correct:      0.500185
Highest Incorrect:   0.786890

=========================================================================================
Comparator:          SorensenDice ngram=4
Accuracy:            98.720948% Total Comparisons Made:   66690
Threshold:           0.630000
False Positives:     468 0.701754%
False Negatives:     385 0.577298%
Lowest Correct:      0.177778
Highest Incorrect:   0.730159

=========================================================================================
Comparator:          Jaccard ngram=2 EqualLengths
Accuracy:            98.731444% Total Comparisons Made:   66690
Threshold:           0.530000
False Positives:     19 0.028490%
False Negatives:     827 1.240066%
Lowest Correct:      0.148649
Highest Incorrect:   0.541667

=========================================================================================
Comparator:          Jaccard ngram=2 EqualLengths
Accuracy:            98.731444% Total Comparisons Made:   66690
Threshold:           0.520000
False Positives:     19 0.028490%
False Negatives:     827 1.240066%
Lowest Correct:      0.148649
Highest Incorrect:   0.541667

=========================================================================================
Comparator:          OverlapCoefficient ngram=2 EqualLengths
Accuracy:            98.731444% Total Comparisons Made:   66690
Threshold:           0.690000
False Positives:     19 0.028490%
False Negatives:     827 1.240066%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          SorensenDice ngram=2 EqualLengths
Accuracy:            98.731444% Total Comparisons Made:   66690
Threshold:           0.690000
False Positives:     19 0.028490%
False Negatives:     827 1.240066%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          OverlapCoefficient ngram=2 EqualLengths
Accuracy:            98.758435% Total Comparisons Made:   66690
Threshold:           0.700000
False Positives:     1 0.001499%
False Negatives:     827 1.240066%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          Jaccard ngram=2 EqualLengths
Accuracy:            98.758435% Total Comparisons Made:   66690
Threshold:           0.540000
False Positives:     1 0.001499%
False Negatives:     827 1.240066%
Lowest Correct:      0.148649
Highest Incorrect:   0.541667

=========================================================================================
Comparator:          SorensenDice ngram=2 EqualLengths
Accuracy:            98.758435% Total Comparisons Made:   66690
Threshold:           0.700000
False Positives:     1 0.001499%
False Negatives:     827 1.240066%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          OverlapCoefficient ngram=4
Accuracy:            98.771930% Total Comparisons Made:   66690
Threshold:           0.700000
False Positives:     306 0.458839%
False Negatives:     513 0.769231%
Lowest Correct:      0.230769
Highest Incorrect:   0.793103

=========================================================================================
Comparator:          OverlapCoefficient ngram=3
Accuracy:            98.777928% Total Comparisons Made:   66690
Threshold:           0.740000
False Positives:     306 0.458839%
False Negatives:     509 0.763233%
Lowest Correct:      0.264151
Highest Incorrect:   0.800000

=========================================================================================
Comparator:          OverlapCoefficient ngram=3
Accuracy:            98.777928% Total Comparisons Made:   66690
Threshold:           0.730000
False Positives:     306 0.458839%
False Negatives:     509 0.763233%
Lowest Correct:      0.264151
Highest Incorrect:   0.800000

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            98.786925% Total Comparisons Made:   66690
Threshold:           0.700000
False Positives:     639 0.958165%
False Negatives:     170 0.254911%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          SorensenDice ngram=4
Accuracy:            98.795921% Total Comparisons Made:   66690
Threshold:           0.640000
False Positives:     306 0.458839%
False Negatives:     497 0.745239%
Lowest Correct:      0.177778
Highest Incorrect:   0.730159

=========================================================================================
Comparator:          SorensenDice ngram=3
Accuracy:            98.807917% Total Comparisons Made:   66690
Threshold:           0.640000
False Positives:     463 0.694257%
False Negatives:     332 0.497826%
Lowest Correct:      0.204380
Highest Incorrect:   0.738462

=========================================================================================
Comparator:          Jaro
Accuracy:            98.810916% Total Comparisons Made:   66690
Threshold:           0.750000
False Positives:     193 0.289399%
False Negatives:     600 0.899685%
Lowest Correct:      0.591939
Highest Incorrect:   0.827712

=========================================================================================
Comparator:          OverlapCoefficient ngram=3
Accuracy:            98.812416% Total Comparisons Made:   66690
Threshold:           0.680000
False Positives:     620 0.929675%
False Negatives:     172 0.257910%
Lowest Correct:      0.264151
Highest Incorrect:   0.800000

=========================================================================================
Comparator:          SmithWatermanGotoh gap=-0.1 match=1 mismatch=-0.25
Accuracy:            98.818414% Total Comparisons Made:   66690
Threshold:           0.750000
False Positives:     610 0.914680%
False Negatives:     178 0.266907%
Lowest Correct:      0.447273
Highest Incorrect:   0.900000

=========================================================================================
Comparator:          SmithWatermanGotoh gap=-0.1 match=1 mismatch=-0.5
Accuracy:            98.818414% Total Comparisons Made:   66690
Threshold:           0.750000
False Positives:     610 0.914680%
False Negatives:     178 0.266907%
Lowest Correct:      0.447273
Highest Incorrect:   0.900000

=========================================================================================
Comparator:          Levenshtein insert=1 replace=2 delete=1
Accuracy:            98.821413% Total Comparisons Made:   66690
Threshold:           0.550000
False Positives:     91 0.136452%
False Negatives:     695 1.042135%
Lowest Correct:      0.081395
Highest Incorrect:   0.648649

=========================================================================================
Comparator:          SorensenDice ngram=2
Accuracy:            98.830409% Total Comparisons Made:   66690
Threshold:           0.650000
False Positives:     490 0.734743%
False Negatives:     290 0.434848%
Lowest Correct:      0.316547
Highest Incorrect:   0.746269

=========================================================================================
Comparator:          Jaro
Accuracy:            98.831909% Total Comparisons Made:   66690
Threshold:           0.740000
False Positives:     320 0.479832%
False Negatives:     459 0.688259%
Lowest Correct:      0.591939
Highest Incorrect:   0.827712

=========================================================================================
Comparator:          OverlapCoefficient ngram=3
Accuracy:            98.833408% Total Comparisons Made:   66690
Threshold:           0.690000
False Positives:     494 0.740741%
False Negatives:     284 0.425851%
Lowest Correct:      0.264151
Highest Incorrect:   0.800000

=========================================================================================
Comparator:          Levenshtein insert=1 replace=2 delete=1
Accuracy:            98.839406% Total Comparisons Made:   66690
Threshold:           0.560000
False Positives:     79 0.118459%
False Negatives:     695 1.042135%
Lowest Correct:      0.081395
Highest Incorrect:   0.648649

=========================================================================================
Comparator:          Levenshtein insert=1 replace=2 delete=1
Accuracy:            98.848403% Total Comparisons Made:   66690
Threshold:           0.570000
False Positives:     73 0.109462%
False Negatives:     695 1.042135%
Lowest Correct:      0.081395
Highest Incorrect:   0.648649

=========================================================================================
Comparator:          Levenshtein insert=1 replace=2 delete=1
Accuracy:            98.849903% Total Comparisons Made:   66690
Threshold:           0.580000
False Positives:     72 0.107962%
False Negatives:     695 1.042135%
Lowest Correct:      0.081395
Highest Incorrect:   0.648649

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            98.858899% Total Comparisons Made:   66690
Threshold:           0.730000
False Positives:     585 0.877193%
False Negatives:     176 0.263908%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          SmithWatermanGotoh gap=-0.25 match=1 mismatch=-0.25
Accuracy:            98.858899% Total Comparisons Made:   66690
Threshold:           0.710000
False Positives:     589 0.883191%
False Negatives:     172 0.257910%
Lowest Correct:      0.322727
Highest Incorrect:   0.886364

=========================================================================================
Comparator:          Jaro
Accuracy:            98.879892% Total Comparisons Made:   66690
Threshold:           0.720000
False Positives:     559 0.838207%
False Negatives:     188 0.281901%
Lowest Correct:      0.591939
Highest Incorrect:   0.827712

=========================================================================================
Comparator:          Jaro
Accuracy:            98.887389% Total Comparisons Made:   66690
Threshold:           0.760000
False Positives:     96 0.143950%
False Negatives:     646 0.968661%
Lowest Correct:      0.591939
Highest Incorrect:   0.827712

=========================================================================================
Comparator:          SmithWatermanGotoh gap=-0.25 match=1 mismatch=-0.25
Accuracy:            98.921877% Total Comparisons Made:   66690
Threshold:           0.720000
False Positives:     547 0.820213%
False Negatives:     172 0.257910%
Lowest Correct:      0.322727
Highest Incorrect:   0.886364

=========================================================================================
Comparator:          SorensenDice ngram=4
Accuracy:            98.939871% Total Comparisons Made:   66690
Threshold:           0.650000
False Positives:     210 0.314890%
False Negatives:     497 0.745239%
Lowest Correct:      0.177778
Highest Incorrect:   0.730159

=========================================================================================
Comparator:          Jaccard ngram=3
Accuracy:            98.947368% Total Comparisons Made:   66690
Threshold:           0.500000
False Positives:     211 0.316389%
False Negatives:     491 0.736242%
Lowest Correct:      0.113821
Highest Incorrect:   0.585366

=========================================================================================
Comparator:          Jaccard ngram=3
Accuracy:            98.948868% Total Comparisons Made:   66690
Threshold:           0.510000
False Positives:     210 0.314890%
False Negatives:     491 0.736242%
Lowest Correct:      0.113821
Highest Incorrect:   0.585366

=========================================================================================
Comparator:          SorensenDice ngram=3
Accuracy:            98.948868% Total Comparisons Made:   66690
Threshold:           0.670000
False Positives:     210 0.314890%
False Negatives:     491 0.736242%
Lowest Correct:      0.204380
Highest Incorrect:   0.738462

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            98.963863% Total Comparisons Made:   66690
Threshold:           0.740000
False Positives:     513 0.769231%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          OverlapCoefficient ngram=3
Accuracy:            98.971360% Total Comparisons Made:   66690
Threshold:           0.720000
False Positives:     339 0.508322%
False Negatives:     347 0.520318%
Lowest Correct:      0.264151
Highest Incorrect:   0.800000

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            98.972859% Total Comparisons Made:   66690
Threshold:           0.710000
False Positives:     511 0.766232%
False Negatives:     174 0.260909%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          SorensenDice ngram=3
Accuracy:            98.972859% Total Comparisons Made:   66690
Threshold:           0.650000
False Positives:     307 0.460339%
False Negatives:     378 0.566802%
Lowest Correct:      0.204380
Highest Incorrect:   0.738462

=========================================================================================
Comparator:          SorensenDice ngram=3
Accuracy:            98.984855% Total Comparisons Made:   66690
Threshold:           0.680000
False Positives:     180 0.269906%
False Negatives:     497 0.745239%
Lowest Correct:      0.204380
Highest Incorrect:   0.738462

=========================================================================================
Comparator:          Jaccard ngram=3
Accuracy:            98.984855% Total Comparisons Made:   66690
Threshold:           0.520000
False Positives:     180 0.269906%
False Negatives:     497 0.745239%
Lowest Correct:      0.113821
Highest Incorrect:   0.585366

=========================================================================================
Comparator:          OverlapCoefficient ngram=2 EqualLengths
Accuracy:            98.998351% Total Comparisons Made:   66690
Threshold:           0.640000
False Positives:     346 0.518818%
False Negatives:     322 0.482831%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          SorensenDice ngram=2 EqualLengths
Accuracy:            98.998351% Total Comparisons Made:   66690
Threshold:           0.640000
False Positives:     346 0.518818%
False Negatives:     322 0.482831%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            99.001350% Total Comparisons Made:   66690
Threshold:           0.500000
False Positives:     334 0.500825%
False Negatives:     332 0.497826%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

=========================================================================================
Comparator:          Jaccard ngram=3
Accuracy:            99.002849% Total Comparisons Made:   66690
Threshold:           0.530000
False Positives:     156 0.233918%
False Negatives:     509 0.763233%
Lowest Correct:      0.113821
Highest Incorrect:   0.585366

=========================================================================================
Comparator:          SorensenDice ngram=3
Accuracy:            99.020843% Total Comparisons Made:   66690
Threshold:           0.690000
False Positives:     156 0.233918%
False Negatives:     497 0.745239%
Lowest Correct:      0.204380
Highest Incorrect:   0.738462

=========================================================================================
Comparator:          OverlapCoefficient ngram=4
Accuracy:            99.020843% Total Comparisons Made:   66690
Threshold:           0.690000
False Positives:     306 0.458839%
False Negatives:     347 0.520318%
Lowest Correct:      0.230769
Highest Incorrect:   0.793103

=========================================================================================
Comparator:          SorensenDice ngram=3 EqualLengths
Accuracy:            99.023842% Total Comparisons Made:   66690
Threshold:           0.630000
False Positives:     319 0.478333%
False Negatives:     332 0.497826%
Lowest Correct:      0.166667
Highest Incorrect:   0.685714

=========================================================================================
Comparator:          OverlapCoefficient ngram=3 EqualLengths
Accuracy:            99.023842% Total Comparisons Made:   66690
Threshold:           0.630000
False Positives:     319 0.478333%
False Negatives:     332 0.497826%
Lowest Correct:      0.166667
Highest Incorrect:   0.685714

=========================================================================================
Comparator:          OverlapCoefficient ngram=4
Accuracy:            99.023842% Total Comparisons Made:   66690
Threshold:           0.680000
False Positives:     339 0.508322%
False Negatives:     312 0.467836%
Lowest Correct:      0.230769
Highest Incorrect:   0.793103

=========================================================================================
Comparator:          OverlapCoefficient ngram=4
Accuracy:            99.041835% Total Comparisons Made:   66690
Threshold:           0.670000
False Positives:     339 0.508322%
False Negatives:     300 0.449843%
Lowest Correct:      0.230769
Highest Incorrect:   0.793103

=========================================================================================
Comparator:          OverlapCoefficient ngram=3
Accuracy:            99.041835% Total Comparisons Made:   66690
Threshold:           0.710000
False Positives:     339 0.508322%
False Negatives:     300 0.449843%
Lowest Correct:      0.264151
Highest Incorrect:   0.800000

=========================================================================================
Comparator:          SorensenDice ngram=2
Accuracy:            99.049333% Total Comparisons Made:   66690
Threshold:           0.660000
False Positives:     334 0.500825%
False Negatives:     300 0.449843%
Lowest Correct:      0.316547
Highest Incorrect:   0.746269

=========================================================================================
Comparator:          OverlapCoefficient ngram=4
Accuracy:            99.065827% Total Comparisons Made:   66690
Threshold:           0.660000
False Positives:     339 0.508322%
False Negatives:     284 0.425851%
Lowest Correct:      0.230769
Highest Incorrect:   0.793103

=========================================================================================
Comparator:          OverlapCoefficient ngram=3
Accuracy:            99.065827% Total Comparisons Made:   66690
Threshold:           0.700000
False Positives:     339 0.508322%
False Negatives:     284 0.425851%
Lowest Correct:      0.264151
Highest Incorrect:   0.800000

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            99.098815% Total Comparisons Made:   66690
Threshold:           0.850000
False Positives:     1 0.001499%
False Negatives:     600 0.899685%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          Levenshtein insert=1 replace=2 delete=1
Accuracy:            99.100315% Total Comparisons Made:   66690
Threshold:           0.500000
False Positives:     422 0.632779%
False Negatives:     178 0.266907%
Lowest Correct:      0.081395
Highest Incorrect:   0.648649

=========================================================================================
Comparator:          SorensenDice ngram=3
Accuracy:            99.106313% Total Comparisons Made:   66690
Threshold:           0.660000
False Positives:     211 0.316389%
False Negatives:     385 0.577298%
Lowest Correct:      0.204380
Highest Incorrect:   0.738462

=========================================================================================
Comparator:          Jaro
Accuracy:            99.118309% Total Comparisons Made:   66690
Threshold:           0.730000
False Positives:     392 0.587794%
False Negatives:     196 0.293897%
Lowest Correct:      0.591939
Highest Incorrect:   0.827712

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.121308% Total Comparisons Made:   66690
Threshold:           0.830000
False Positives:     1 0.001499%
False Negatives:     585 0.877193%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          Levenshtein insert=1 replace=2 delete=1
Accuracy:            99.136302% Total Comparisons Made:   66690
Threshold:           0.510000
False Positives:     398 0.596791%
False Negatives:     178 0.266907%
Lowest Correct:      0.081395
Highest Incorrect:   0.648649

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            99.145299% Total Comparisons Made:   66690
Threshold:           0.510000
False Positives:     238 0.356875%
False Negatives:     332 0.497826%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

=========================================================================================
Comparator:          SorensenDice ngram=2
Accuracy:            99.145299% Total Comparisons Made:   66690
Threshold:           0.670000
False Positives:     238 0.356875%
False Negatives:     332 0.497826%
Lowest Correct:      0.316547
Highest Incorrect:   0.746269

=========================================================================================
Comparator:          Levenshtein insert=1 replace=2 delete=1
Accuracy:            99.151297% Total Comparisons Made:   66690
Threshold:           0.540000
False Positives:     175 0.262408%
False Negatives:     391 0.586295%
Lowest Correct:      0.081395
Highest Incorrect:   0.648649

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            99.152797% Total Comparisons Made:   66690
Threshold:           0.750000
False Positives:     387 0.580297%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            99.164792% Total Comparisons Made:   66690
Threshold:           0.530000
False Positives:     181 0.271405%
False Negatives:     376 0.563803%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

=========================================================================================
Comparator:          Levenshtein insert=1 replace=2 delete=1
Accuracy:            99.164792% Total Comparisons Made:   66690
Threshold:           0.520000
False Positives:     379 0.568301%
False Negatives:     178 0.266907%
Lowest Correct:      0.081395
Highest Incorrect:   0.648649

=========================================================================================
Comparator:          SorensenDice ngram=2
Accuracy:            99.164792% Total Comparisons Made:   66690
Threshold:           0.690000
False Positives:     181 0.271405%
False Negatives:     376 0.563803%
Lowest Correct:      0.316547
Highest Incorrect:   0.746269

=========================================================================================
Comparator:          SorensenDice ngram=2
Accuracy:            99.185785% Total Comparisons Made:   66690
Threshold:           0.680000
False Positives:     211 0.316389%
False Negatives:     332 0.497826%
Lowest Correct:      0.316547
Highest Incorrect:   0.746269

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            99.185785% Total Comparisons Made:   66690
Threshold:           0.520000
False Positives:     211 0.316389%
False Negatives:     332 0.497826%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            99.190283% Total Comparisons Made:   66690
Threshold:           0.540000
False Positives:     157 0.235418%
False Negatives:     383 0.574299%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

=========================================================================================
Comparator:          SorensenDice ngram=2
Accuracy:            99.190283% Total Comparisons Made:   66690
Threshold:           0.700000
False Positives:     157 0.235418%
False Negatives:     383 0.574299%
Lowest Correct:      0.316547
Highest Incorrect:   0.746269

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            99.191783% Total Comparisons Made:   66690
Threshold:           0.550000
False Positives:     156 0.233918%
False Negatives:     383 0.574299%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

=========================================================================================
Comparator:          SorensenDice ngram=2
Accuracy:            99.191783% Total Comparisons Made:   66690
Threshold:           0.710000
False Positives:     156 0.233918%
False Negatives:     383 0.574299%
Lowest Correct:      0.316547
Highest Incorrect:   0.746269

=========================================================================================
Comparator:          OverlapCoefficient ngram=3 EqualLengths
Accuracy:            99.197781% Total Comparisons Made:   66690
Threshold:           0.640000
False Positives:     157 0.235418%
False Negatives:     378 0.566802%
Lowest Correct:      0.166667
Highest Incorrect:   0.685714

=========================================================================================
Comparator:          SorensenDice ngram=3 EqualLengths
Accuracy:            99.197781% Total Comparisons Made:   66690
Threshold:           0.640000
False Positives:     157 0.235418%
False Negatives:     378 0.566802%
Lowest Correct:      0.166667
Highest Incorrect:   0.685714

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.208277% Total Comparisons Made:   66690
Threshold:           0.720000
False Positives:     352 0.527815%
False Negatives:     176 0.263908%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          OverlapCoefficient ngram=2 EqualLengths
Accuracy:            99.226271% Total Comparisons Made:   66690
Threshold:           0.650000
False Positives:     184 0.275903%
False Negatives:     332 0.497826%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          SorensenDice ngram=2 EqualLengths
Accuracy:            99.226271% Total Comparisons Made:   66690
Threshold:           0.650000
False Positives:     184 0.275903%
False Negatives:     332 0.497826%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          Levenshtein insert=1 replace=2 delete=1
Accuracy:            99.227770% Total Comparisons Made:   66690
Threshold:           0.530000
False Positives:     337 0.505323%
False Negatives:     178 0.266907%
Lowest Correct:      0.081395
Highest Incorrect:   0.648649

=========================================================================================
Comparator:          SorensenDice ngram=2 EqualLengths
Accuracy:            99.266757% Total Comparisons Made:   66690
Threshold:           0.660000
False Positives:     157 0.235418%
False Negatives:     332 0.497826%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          OverlapCoefficient ngram=2 EqualLengths
Accuracy:            99.266757% Total Comparisons Made:   66690
Threshold:           0.660000
False Positives:     157 0.235418%
False Negatives:     332 0.497826%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          Jaccard ngram=2 EqualLengths
Accuracy:            99.266757% Total Comparisons Made:   66690
Threshold:           0.500000
False Positives:     157 0.235418%
False Negatives:     332 0.497826%
Lowest Correct:      0.148649
Highest Incorrect:   0.541667

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            99.296746% Total Comparisons Made:   66690
Threshold:           0.760000
False Positives:     291 0.436347%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.310241% Total Comparisons Made:   66690
Threshold:           0.820000
False Positives:     1 0.001499%
False Negatives:     459 0.688259%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            99.332734% Total Comparisons Made:   66690
Threshold:           0.570000
False Positives:     18 0.026991%
False Negatives:     427 0.640276%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.361224% Total Comparisons Made:   66690
Threshold:           0.770000
False Positives:     35 0.052482%
False Negatives:     391 0.586295%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.386715% Total Comparisons Made:   66690
Threshold:           0.730000
False Positives:     231 0.346379%
False Negatives:     178 0.266907%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          SorensenDice ngram=2
Accuracy:            99.386715% Total Comparisons Made:   66690
Threshold:           0.720000
False Positives:     18 0.026991%
False Negatives:     391 0.586295%
Lowest Correct:      0.316547
Highest Incorrect:   0.746269

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            99.386715% Total Comparisons Made:   66690
Threshold:           0.560000
False Positives:     18 0.026991%
False Negatives:     391 0.586295%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.392713% Total Comparisons Made:   66690
Threshold:           0.780000
False Positives:     14 0.020993%
False Negatives:     391 0.586295%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          OverlapCoefficient ngram=2 EqualLengths
Accuracy:            99.407707% Total Comparisons Made:   66690
Threshold:           0.680000
False Positives:     19 0.028490%
False Negatives:     376 0.563803%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          SorensenDice ngram=2 EqualLengths
Accuracy:            99.407707% Total Comparisons Made:   66690
Threshold:           0.680000
False Positives:     19 0.028490%
False Negatives:     376 0.563803%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.410706% Total Comparisons Made:   66690
Threshold:           0.810000
False Positives:     2 0.002999%
False Negatives:     391 0.586295%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.410706% Total Comparisons Made:   66690
Threshold:           0.800000
False Positives:     2 0.002999%
False Negatives:     391 0.586295%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.410706% Total Comparisons Made:   66690
Threshold:           0.790000
False Positives:     2 0.002999%
False Negatives:     391 0.586295%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            99.431699% Total Comparisons Made:   66690
Threshold:           0.770000
False Positives:     201 0.301395%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            99.467686% Total Comparisons Made:   66690
Threshold:           0.780000
False Positives:     177 0.265407%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          Jaccard ngram=2 EqualLengths
Accuracy:            99.473684% Total Comparisons Made:   66690
Threshold:           0.510000
False Positives:     19 0.028490%
False Negatives:     332 0.497826%
Lowest Correct:      0.148649
Highest Incorrect:   0.541667

=========================================================================================
Comparator:          SorensenDice ngram=2 EqualLengths
Accuracy:            99.473684% Total Comparisons Made:   66690
Threshold:           0.670000
False Positives:     19 0.028490%
False Negatives:     332 0.497826%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          OverlapCoefficient ngram=2 EqualLengths
Accuracy:            99.473684% Total Comparisons Made:   66690
Threshold:           0.670000
False Positives:     19 0.028490%
False Negatives:     332 0.497826%
Lowest Correct:      0.258824
Highest Incorrect:   0.702703

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            99.505173% Total Comparisons Made:   66690
Threshold:           0.790000
False Positives:     152 0.227920%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            99.514170% Total Comparisons Made:   66690
Threshold:           0.800000
False Positives:     146 0.218923%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.557655% Total Comparisons Made:   66690
Threshold:           0.740000
False Positives:     117 0.175439%
False Negatives:     178 0.266907%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            99.586145% Total Comparisons Made:   66690
Threshold:           0.810000
False Positives:     98 0.146949%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.608637% Total Comparisons Made:   66690
Threshold:           0.750000
False Positives:     83 0.124456%
False Negatives:     178 0.266907%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            99.628130% Total Comparisons Made:   66690
Threshold:           0.840000
False Positives:     2 0.002999%
False Negatives:     246 0.368871%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.634128% Total Comparisons Made:   66690
Threshold:           0.760000
False Positives:     66 0.098965%
False Negatives:     178 0.266907%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            99.712101% Total Comparisons Made:   66690
Threshold:           0.820000
False Positives:     14 0.020993%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          JaroWinkler
Accuracy:            99.730094% Total Comparisons Made:   66690
Threshold:           0.830000
False Positives:     2 0.002999%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348
```
