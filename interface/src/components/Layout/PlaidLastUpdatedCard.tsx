import { formatDistanceToNow } from 'date-fns';

import MSpan from '@monetr/interface/components/MSpan';
import Typography from '@monetr/interface/components/Typography';
import { useLink } from '@monetr/interface/hooks/useLink';

interface PlaidLastUpdatedCardProps {
  linkId: string | null;
}

export default function PlaidLastUpdatedCard(props: PlaidLastUpdatedCardProps): JSX.Element {
  const link = useLink(props?.linkId);

  if (!link?.data?.plaidLink) {
    return null;
  }

  const lastUpdateString = link?.data?.plaidLink?.lastSuccessfulUpdate
    ? formatDistanceToNow(link.data.plaidLink.lastSuccessfulUpdate)
    : 'Never';

  const lastAttemptString = link?.data?.plaidLink?.lastAttemptedUpdate
    ? formatDistanceToNow(link.data.plaidLink.lastAttemptedUpdate)
    : 'Never';

  return (
    <div className='p-2 group border-[thin] dark:border-dark-monetr-border rounded-lg w-full ease-in-out transition-[height] h-16 lg:h-10 lg:hover:h-16 hover:delay-0 delay-500 hover:dark:border-dark-monetr-border-string'>
      <Typography color='subtle' size='sm'>
        Last Updated: {lastUpdateString} ago
      </Typography>
      <MSpan
        className='transition-opacity opacity-100 lg:opacity-0 lg:group-hover:opacity-100 delay-500 group-hover:delay-0'
        color='subtle'
        size='sm'
      >
        Last Attempt: {lastAttemptString} ago
      </MSpan>
    </div>
  );
}
