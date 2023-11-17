# Recurring

This package contains the code necessary for detecting recurring transactions.

## Compare

In order to compare two transactions to determine whether or not they are the same or similar a text comparison 
algorithm is used. This algorithm is chosen based on tests to analyze its accuracy against a dataset with a variety 
of different parameters. At the time of writing this the `JaroWinkler` text comparison algorithm with strings 
adjusted to be equal lengths and a success threshold of `0.75` is the best performing configuration.

The top several configurations are listed below, from least accurate to most accurate:

```text
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
Comparator:          SorensenDice ngram=3
Accuracy:            98.972859% Total Comparisons Made:   66690
Threshold:           0.650000
False Positives:     307 0.460339%
False Negatives:     378 0.566802%
Lowest Correct:      0.204380
Highest Incorrect:   0.738462

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            99.001350% Total Comparisons Made:   66690
Threshold:           0.500000
False Positives:     334 0.500825%
False Negatives:     332 0.497826%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

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
Comparator:          JaroWinkler
Accuracy:            99.152797% Total Comparisons Made:   66690
Threshold:           0.750000
False Positives:     387 0.580297%
False Negatives:     178 0.266907%
Lowest Correct:      0.591939
Highest Incorrect:   0.855348

=========================================================================================
Comparator:          Jaccard ngram=2
Accuracy:            99.191783% Total Comparisons Made:   66690
Threshold:           0.550000
False Positives:     156 0.233918%
False Negatives:     383 0.574299%
Lowest Correct:      0.188034
Highest Incorrect:   0.595238

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
Comparator:          Jaccard ngram=2 EqualLengths
Accuracy:            99.266757% Total Comparisons Made:   66690
Threshold:           0.500000
False Positives:     157 0.235418%
False Negatives:     332 0.497826%
Lowest Correct:      0.148649
Highest Incorrect:   0.541667

=========================================================================================
Comparator:          JaroWinkler EqualLengths
Accuracy:            99.410706% Total Comparisons Made:   66690
Threshold:           0.800000
False Positives:     2 0.002999%
False Negatives:     391 0.586295%
Lowest Correct:      0.500185
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
Accuracy:            99.608637% Total Comparisons Made:   66690
Threshold:           0.750000
False Positives:     83 0.124456%
False Negatives:     178 0.266907%
Lowest Correct:      0.500185
Highest Incorrect:   0.855348
```
