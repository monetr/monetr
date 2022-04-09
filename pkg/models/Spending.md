# Spending calculations

This document is just some notes needed for the future. I want to try to implement a better way to calculate
contributions to spending objects.

---

1. Take the `nextContributionDate` and `nextRecurrence` date. If either are in the past relative to `now` then increment
   ones that are in the past to the first occurrence after `now`.
2. ```go
   switch {
   case nextContributionDate.After(nextRecurrence):
     // If the next recurrence will happen before the next contribution then calculate the number of recurrences total
     // that will happen before the next contribution.
     nextContributionAmount = (targetAmount * numberOfRecurrences) - currentAmount;
   case nextRecurrence.After(nextContributionDate):
     // The next contribution will happen before the next recurrence. Calculate the number of contributions that will
     // occur before the next recurrence to determine the contribution amount needed.
     nextContributionAmount = (targetAmount - currentAmount) / numberOfContributions;
   case nextRecurrence.Equal(nextContributionDate):
     // The next recurrence is on the same day as the next contribution. Allocate however much is missing.
     nextContributionAmount = targetAmount - currentAmount;
   }
   ```

This _should_ result in the correct contribution amounts regardless of scenario, and should be a far cleaner layout for
the code than it currently is. But I have not yet tested this and it is only a theory at the moment.
