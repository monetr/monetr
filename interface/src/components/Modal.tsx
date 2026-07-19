import React from 'react';

import Typography, { type TypographyProps } from '@monetr/interface/components/Typography';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Modal.module.scss';

export interface ModalProps {
  open: boolean;
  onClose?: () => void;
  children: React.ReactNode;
  className?: string;
}

export type ModalRef = HTMLDivElement;

const Modal = React.forwardRef<ModalRef, ModalProps>((props, ref) => {
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

export function ModalTitle(props: TypographyProps): React.JSX.Element {
  return <Typography size='xl' weight='bold' {...props} className={mergeClasses(props.className)} />;
}

export function ModalDescription(props: TypographyProps): React.JSX.Element {
  return (
    <Typography color='subtle' weight='normal' {...props} className={mergeClasses(styles.heading, props.className)} />
  );
}

export function ModalContent({ children }: React.PropsWithChildren): React.JSX.Element {
  return <div className={styles.content}>{children}</div>;
}

export function ModalActions({ children }: React.PropsWithChildren): React.JSX.Element {
  return <div className={styles.actions}>{children}</div>;
}

export default Modal;
