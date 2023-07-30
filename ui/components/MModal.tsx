/* eslint-disable max-len */
import React from 'react';

import { ReactElement } from './types';

export interface MModalProps {
  open: boolean;
  onClose?: () => void;
  children: ReactElement;
}

export type MModalRef = HTMLDivElement;

const MModal = React.forwardRef<MModalRef, MModalProps>((props, ref) => {
  if (!props.open) {
    return null;
  }

  return (
    <div className='top-0 left-0 absolute z-40 w-screen h-screen overflow-hidden' role='dialog' aria-modal='true'>
      <div className='fixed inset-0 bg-dark-monetr-background bg-opacity-50 transition-opacity backdrop-blur-sm backdrop-brightness-50' />
      <div ref={ ref } className="fixed inset-0 z-10 overflow-y-hidden h-screen max-h-screen">
        <div className="flex justify-center p-4 items-center sm:p-0 min-h-full">
          <div className="relative transform overflow-hidden rounded-lg dark:bg-dark-monetr-background shadow-xl sm:w-full sm:max-w-lg p-2">
            { props.children }
          </div>
        </div>
      </div>
    </div>
  );
});

export default MModal;

