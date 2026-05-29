import { Landmark } from 'lucide-react';

import { useInstitution } from '@monetr/interface/hooks/useInstitution';
import type Link from '@monetr/interface/models/Link';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './InstitutionLogo.module.scss';

interface PlaidInstitutionLogoProps {
  link?: Link;
  className?: string;
}

export default function PlaidInstitutionLogo(props: PlaidInstitutionLogoProps): JSX.Element {
  const { data: institution } = useInstitution(props.link?.plaidLink?.institutionId);

  if (!institution?.logo) {
    return (
      <Landmark
        className={mergeClasses(styles.iconFallback, props.className)}
        data-testid={`bank-sidebar-item-${props.link?.linkId}-logo-missing`}
      />
    );
  }

  return (
    <img
      alt={institution.name}
      className={props.className}
      data-testid={`bank-sidebar-item-${props.link?.linkId}-logo`}
      src={`data:image/png;base64,${institution.logo}`}
    />
  );
}
