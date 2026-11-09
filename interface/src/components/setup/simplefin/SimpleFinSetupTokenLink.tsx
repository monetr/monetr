import { ExternalLink } from 'lucide-react';

import { textLinkVariants } from '@monetr/interface/components/TextLink';

import styles from './SimpleFinSetupTokenLink.module.scss';

export default function SimpleFinSetupTokenLink(): React.JSX.Element {
  return (
    <a
      className={textLinkVariants({ size: 'sm' })}
      href='https://bridge.simplefin.org/simplefin/create'
      rel='noopener noreferrer'
      tabIndex={-1}
      target='_blank'
    >
      Get A Setup Token
      <ExternalLink className={styles.externalLinkIcon} />
    </a>
  );
}
