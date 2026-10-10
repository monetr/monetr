import { Fragment } from 'react';
import { Clock, Plug, TriangleAlert } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Typography from '@monetr/interface/components/Typography';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { showUpdatePlaidAccountOverlay } from '@monetr/interface/modals/UpdatePlaidAccountOverlay';
import type Link from '@monetr/interface/models/Link';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './LinkAlerts.module.scss';

export interface LinkAlertsProps {
  link: Link;
}

export default function LinkAlerts(props: LinkAlertsProps): React.JSX.Element | null {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocale();

  if (!props.link.getIsPlaid()) {
    return null;
  }

  function reauthenticate() {
    showUpdatePlaidAccountOverlay({
      link: props.link,
    });
  }

  function updateAccountSelection() {
    showUpdatePlaidAccountOverlay({
      link: props.link,
      updateAccountSelection: true,
    });
  }

  const expirationDate = props.link.plaidLink?.expirationDate;

  return (
    <Fragment>
      {props.link.getIsPendingExpiration() && (
        <div className={styles.alert} data-variant='warning' role='status'>
          <Clock className={styles.icon} />
          <div className={styles.text}>
            <Typography color='emphasis' component='p' weight='semibold'>
              {props.link.getName()} access expires
              {expirationDate && locale && ` on ${formatDate(expirationDate, inTimezone, locale, DateLength.Medium)}`}
            </Typography>
            <Typography component='p'>Reauthenticate before then to keep transactions syncing.</Typography>
          </div>
          <Button onClick={reauthenticate} variant='outlined'>
            Reauthenticate
          </Button>
        </div>
      )}
      {props.link.getIsError() && (
        <div className={styles.alert} data-variant='error' role='alert'>
          <TriangleAlert className={styles.icon} />
          <div className={styles.text}>
            <Typography color='emphasis' component='p' weight='semibold'>
              {props.link.getName()} needs you to sign in again
            </Typography>
            <Typography component='p'>Reauthenticate to start syncing again.</Typography>
          </div>
          <Button onClick={reauthenticate} variant='destructive'>
            Reauthenticate
          </Button>
        </div>
      )}
      {props.link.getCanUpdateAccountSelection() && (
        <div className={styles.alert} data-variant='brand' role='status'>
          <Plug className={styles.icon} />
          <div className={styles.text}>
            <Typography color='emphasis' component='p' weight='semibold'>
              Plaid found new accounts at {props.link.getName()}
            </Typography>
            <Typography component='p'>
              Pick which ones monetr should see. Your current accounts stay as they are.
            </Typography>
          </div>
          <Button onClick={updateAccountSelection} variant='primary'>
            Update Account Selection
          </Button>
        </div>
      )}
    </Fragment>
  );
}
