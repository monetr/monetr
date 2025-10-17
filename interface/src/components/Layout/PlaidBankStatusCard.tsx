import React from 'react';

import MSpan from '@monetr/interface/components/MSpan';
import PlaidInstitutionLogo from '@monetr/interface/components/Plaid/InstitutionLogo';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import { useInstitution } from '@monetr/interface/hooks/useInstitution';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

/**
 * PlaidBankStatusCard automatically shows the institution health of the currently selected bank account. If the
 * currently selected bank account is **not** a Plaid link, then this will return null.
 */
export default function PlaidBankStatusCard(): JSX.Element {
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
    additionalClasses = 'grayscale';
  }

  return (
    <div className='p-2 group border-[thin] border-dark-monetr-border rounded-lg w-full flex gap-2'>
      <PlaidInstitutionLogo link={link} className={mergeTailwind('w-6 h-6', additionalClasses)} />
      <MSpan size='sm' color='subtle'>
        {status}
      </MSpan>
    </div>
  );
}
