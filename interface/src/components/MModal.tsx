import React from 'react';

import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Modal.module.scss';

export interface MModalProps {
  open: boolean;
  onClose?: () => void;
  children: React.ReactNode;
  className?: string;
}

export type MModalRef = HTMLDivElement;

/**
 * @deprecated Use {@link import('@monetr/interface/components/Modal').default} instead. This component doesn't forward
 * `className`, so callers can't style it. Removal planned for the v2 UI pass.
 * @see {@link import('@monetr/interface/components/Modal').default}
 */
const MModal = React.forwardRef<MModalRef, MModalProps>((props, ref) => {
  if (!props.open) {
    return null;
  }

  return (
    <div aria-modal='true' className={styles.modalRoot} role='dialog'>
      <div className={styles.modalBackdrop} />
      <div className={styles.modalContainer} ref={ref}>
        <div className={styles.modalWindowWrapper}>
          <div className={mergeClasses(styles.modalWindow, props.className)}>{props.children}</div>
        </div>
      </div>
    </div>
  );
});

export default MModal;
