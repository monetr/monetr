import { Link } from 'react-router-dom';

import PlaidInstitutionLogo from '@monetr/interface/components/Plaid/InstitutionLogo';
import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import { useBankAccounts } from '@monetr/interface/hooks/useBankAccounts';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import type MonetrLink from '@monetr/interface/models/Link';
import sortAccounts from '@monetr/interface/util/sortAccounts';

import styles from './BankSidebarItem.module.scss';

interface BankSidebarItemProps {
  link: MonetrLink;
}

export default function BankSidebarItem({ link }: BankSidebarItemProps): JSX.Element {
  const selectBankAccount = useSelectedBankAccount();
  const { data: bankAccounts } = useBankAccounts();
  const active = selectBankAccount.data?.linkId === link.linkId;

  const destinationBankAccounts = sortAccounts(bankAccounts?.filter(bankAccount => bankAccount.linkId === link.linkId));

  const destinationBankAccount = destinationBankAccounts.length > 0 ? destinationBankAccounts[0] : null;

  const LinkWarningIndicator = () => {
    const isWarning = link.getIsError() || link.getIsPendingExpiration();
    if (!isWarning) {
      return null;
    }

    return (
      <span className={styles.statusIndicator}>
        <span className={`${styles.statusPing} ${styles.warningPing}`} />
        <span className={`${styles.statusDot} ${styles.warningDot}`} />
      </span>
    );
  };

  const LinkRevokedIndicator = () => {
    const isBad = link.getIsPlaid() && link.getIsRevoked();
    if (!isBad) {
      return null;
    }

    return (
      <span className={styles.statusIndicator}>
        <span className={`${styles.statusPing} ${styles.revokedPing}`} />
        <span className={`${styles.statusDot} ${styles.revokedDot}`} />
      </span>
    );
  };

  let linkPath = `/bank/${destinationBankAccount?.bankAccountId}/transactions`;
  // If the link has no non-archived bank accounts then instead redirect to the link details page.
  if (bankAccounts?.filter(b => b.linkId === link.linkId).length === 0) {
    linkPath = `/link/${link.linkId}/details`;
  }

  let tooltip: string = link.getName();
  if (link.getIsPlaid()) {
    if (link.getIsError()) {
      tooltip = `${tooltip} (Error)`;
    } else if (link.getIsPendingExpiration()) {
      tooltip = `${tooltip} (Pending Expiration)`;
    } else if (link.getIsRevoked()) {
      tooltip = `${tooltip} (Disconnected)`;
    }
  }

  return (
    <Tooltip delayDuration={100}>
      <TooltipTrigger
        className={styles.root}
        data-testid={`bank-sidebar-item-${link.linkId}`}
      >
        <div className={styles.indicator} data-active={String(active)} />
        <Link className={styles.link} to={linkPath}>
          <PlaidInstitutionLogo link={link} />
          <LinkWarningIndicator />
          <LinkRevokedIndicator />
        </Link>
      </TooltipTrigger>
      <TooltipContent side='right'>{tooltip}</TooltipContent>
    </Tooltip>
  );
}
