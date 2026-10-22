import { Plus } from 'lucide-react';

import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './FloatingActionButton.module.scss';

export interface FloatingActionButtonProps {
  onClick: () => void;
  className?: string;
}

export default function FloatingActionButton(props: FloatingActionButtonProps): React.JSX.Element {
  return (
    <button
      className={mergeClasses(styles.floatingActionButton, props.className)}
      onClick={props.onClick}
      type='button'
    >
      <Plus className={styles.floatingActionButtonIcon} />
    </button>
  );
}
