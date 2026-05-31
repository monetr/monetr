import PlaidInstitutionLogo from '@monetr/interface/components/Plaid/InstitutionLogo';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import { useInstitution } from '@monetr/interface/hooks/useInstitution';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './PlaidBankStatusCard.module.scss';

/**
 * PlaidBankStatusCard automatically shows the institution health of the currently selected bank account. If the
 * currently selected bank account is **not** a Plaid link, then this will return null.
 */
export default function PlaidBankStatusCard(): JSX.Element | null {
  const { data: link } = useCurrentLink();
  const { data: institution } = useInstitution(link?.plaidLink?.institutionId);

  if (!link?.plaidLink) {
    return null;
  }

  let status = 'Institution is healthy';
  let additionalClasses = '';
  if (institution?.status?.transactions_updates?.status !== 'HEALTHY') {
    status = 'Updates may be delayed';
  }
  if (institution?.status?.transactions_updates?.breakdown?.refresh_interval === 'DELAYED') {
    status = 'Updates may be delayed';
  }
  if (institution?.status?.transactions_updates?.breakdown?.refresh_interval === 'STOPPED') {
    status = 'Automatic updates stoppped';
    additionalClasses = styles.grayscale;
  }

  return (
    <div className={styles.card}>
      <PlaidInstitutionLogo className={mergeClasses(styles.logo, additionalClasses)} link={link} />
      <Typography color='subtle' size='sm'>
        {status}
      </Typography>
    </div>
  );
}
