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

