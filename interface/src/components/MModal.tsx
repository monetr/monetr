/* eslint-disable max-len */
import React from 'react';

import { ReactElement } from './types';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MModalProps {
  open: boolean;
  onClose?: () => void;
  children: ReactElement;
  className?: string;
}

export type MModalRef = HTMLDivElement;

const MModal = React.forwardRef<MModalRef, MModalProps>((props, ref) => {
  if (!props.open) {
    return null;
  }

  const classNames = mergeTailwind(
    'dark:bg-dark-monetr-background',
    'overflow-auto',
    'p-2',
    'relative',
    'rounded-none',
    'sm:rounded-lg',
    'shadow-xl',
    'h-screen',
    'sm:h-auto',
    'w-screen',
    'sm:w-full',
    'sm:max-w-lg',
    'transform',
    props.className,
  );

  return (
    <div className='top-0 left-0 absolute z-50 w-screen h-screen overflow-hidden' role='dialog' aria-modal='true'>
      <div className='hidden sm:inline sm:fixed inset-0 bg-dark-monetr-background bg-opacity-50 transition-opacity backdrop-blur-sm backdrop-brightness-50' />
      <div ref={ ref } className='fixed inset-0 z-10 overflow-y-hidden h-screen max-h-screen'>
        <div className='h-auto flex justify-center p-0 sm:p-4 items-center min-h-full'>
          <div className={ classNames }>
            { props.children }
          </div>
        </div>
      </div>
    </div>
  );
});

export default MModal;

