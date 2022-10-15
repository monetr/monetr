# Adding A New Funding Schedule

When you initially connect monetr and Plaid you will not have any funding schedules created yet. You'll want to create a
funding schedule to represent when you get paid. This will help monetr put aside funds for you each paycheck.

Click **Create A Funding Schedule** to get started.

![No Funding Schedules](assets/no_funding_schedules.png)

If you have already created a funding schedule and would like to add another one you can use the floating **Plus** 
button in the bottom right-hand corner of the application.

![Create funding schedule fab button](assets/fab_icon.png)

A dialog should appear giving you a way to tell monetr about your funding schedule.

![Create a funding schedule dialog](assets/create_funding_schedule_dialog.png)

## Steps

The sections below outline each piece of information that monetr needs in order to make use of funding schedules for 
your budgeting.

### Name

A funding schedule's name must be unique a single bank account. If you aren't sure what to name it, we recommend 
something simple like `Payday`.

![Funding schedule name](assets/funding_name.png)

### Next contribution

monetr needs to know when you get paid next for this funding schedule. But if you are trying to fund your budgets 
outside your paycheck then this is really just the next time you want to contribute to your budgets.

![Next funding contribution](assets/paid_next.png)

![Next funding date picker](assets/funding_date_picker.png)

The date picker will default to tomorrow when you open it.

### Frequency

Once you have picked the next funding date, you need to tell monetr how often you get paid. Depending on the day you 
select monetr will give you a few options. (If none of these match your pay frequency please let us know by 
contacting us!)

![Funding frequency](assets/frequency.png)

### Additional Options

If your pay schedule is such that a "pay day" would land on a weekend, but the pay check would have been deposited the
previous business day. (Example; you get paid on the 15th and last day of each month, but one of those days falls on 
a Saturday. Instead, your check is deposited on the day prior on Friday.) You can enable **Exclude Weekends** which will
correct the funding schedule to be the previous business day if the funding schedule falls on a weekend.

![Funding schedule advanced options](assets/funding_options.png)

??? note

    This is still an experimental feature, and while it has been tested there may be some oddities that we
    have not observed yet. If you run into any issues please reach out!

## You're Done!

You should now have a funding schedule created :tada:.

![Successfully created funding schedule](assets/done.png)
