import React from 'react';
import { Landmark } from 'lucide-react';

import { useInstitution } from '@monetr/interface/hooks/institutions';
import Link from '@monetr/interface/models/Link';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

interface PlaidInstitutionLogoProps {
  link?: Link;
  className?: string;
}

export default function PlaidInstitutionLogo(props: PlaidInstitutionLogoProps): JSX.Element {
  const { result: institution } = useInstitution(props.link?.plaidLink?.institutionId);

  if (!institution?.logo) {
    return (
      <Landmark
        data-testid={ `bank-sidebar-item-${props.link?.linkId}-logo-missing` }
        className= { mergeTailwind('text-blue-500', props.className) }
      />
    );
  }

  return (
    <img
      data-testid={ `bank-sidebar-item-${props.link?.linkId}-logo` }
      src={ `data:image/png;base64,${institution.logo}` }
      className= { props.className }
    />
  );
}
