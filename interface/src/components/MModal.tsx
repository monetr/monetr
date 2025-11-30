import React from 'react';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Modal.module.scss';

export interface MModalProps {
  open: boolean;
  onClose?: () => void;
  children: React.ReactNode;
  className?: string;
}

export type MModalRef = HTMLDivElement;

const MModal = React.forwardRef<MModalRef, MModalProps>((props, ref) => {
  if (!props.open) {
    return null;
  }

  return (
    <div aria-modal='true' className={styles.modalRoot} role='dialog'>
      <div className={styles.modalBackdrop} />
      <div className={styles.modalContainer} ref={ref}>
        <div className={styles.modalWindowWrapper}>
          <div className={mergeTailwind(styles.modalWindow, props.className)}>{props.children}</div>
        </div>
      </div>
    </div>
  );
});

export default MModal;
