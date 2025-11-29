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
    <div aria-modal='true' className='top-0 left-0 absolute z-50 w-screen h-screen overflow-hidden' role='dialog'>
      <div className='hidden sm:inline sm:fixed inset-0 bg-dark-monetr-background bg-opacity-50 transition-opacity backdrop-blur-sm backdrop-brightness-50' />
      <div className='fixed inset-0 z-10 overflow-y-hidden h-screen max-h-screen' ref={ref}>
        <div className='h-auto flex justify-center p-0 sm:p-4 items-center min-h-full'>
          <div className={mergeTailwind(styles.modalWindow, props.className)}>{props.children}</div>
        </div>
      </div>
    </div>
  );
});

export default MModal;
