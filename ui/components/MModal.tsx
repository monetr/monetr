/* eslint-disable max-len */
import React from 'react';

import { ReactElement } from './types';

export interface MModalProps {
  open: boolean;
  onClose?: () => void;
  children: ReactElement;
  ref?: React.Ref<HTMLDivElement>;
}

export default function MModal(props: MModalProps): JSX.Element {
  if (!props.open) {
    return null;
  }

  return (
    <div className='relative z-40' role='dialog' aria-modal='true' ref={ props?.ref }>
      <div className='fixed inset-0 bg-dark-monetr-background-emphasis bg-opacity-50 transition-opacity backdrop-blur-sm' />
      <div className="fixed inset-0 z-10 overflow-y-auto">
        <div className="flex justify-center p-4 items-center sm:p-0 min-h-full">
          <div className="relative transform overflow-hidden rounded-lg dark:bg-dark-monetr-background shadow-xl sm:w-full sm:max-w-lg p-2">
            { props.children }
          </div>
        </div>
      </div>
    </div>
  );
}
