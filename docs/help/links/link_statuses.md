# Link Statuses

Links with external software can have a variety of statuses. Manual links typically only have a status of "setup" 
and should always show as a green circle in the UI. But when a link relies on data flow from an integration in order 
to keep monetr up to date, those links can sometimes fail or become degraded.

## Plaid Link Status

Plaid links can have a few different statuses. But the primary ones are as follows:

- `Healthy`: monetr is receiving updates from Plaid, transaction and balance data might be a few hours delayed but 
   should generally be accurate.
- `Degraded`: In the UI this will show as a yellow circle next to the Plaid icon. With a message
   > Updates might be delayed
   
   Plaid is still sending data for this link, but data might be significantly delayed. It is perfectly normal for 
   links to go in and out of this status over time.
- `Down`: The institution is unavailable via Plaid. This can happen if the bank themselves are having problems or if 
   Plaid is offline.
- `Error`: Plaid is no longer sending monetr updates from your financial institution. This can happen for a variety 
  of reasons but is usually resolved by re-authenticating your link.

If your link is missing transactions that are more than 2 days old and your link is not in a "Down" or "Error" state,
please reach out to monetr support, so we may help resolve the issue.

**Note**: Please do not remove a link in an attempt to resolve an issue with it. Removing links deletes all of the data 
associated with it and can cause other side effects. Only remove a link if you have no intention of re-adding it.
