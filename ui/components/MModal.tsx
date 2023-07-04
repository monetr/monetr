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
    <div className='relative z-40' role='dialog' aria-modal='true'>
      <div className='fixed inset-0 bg-zinc-700 bg-opacity-75 transition-opacity' />
      <div className="fixed inset-0 z-10 overflow-y-auto">
        <div className="flex justify-center p-4 items-center sm:p-0 min-h-full">
          <div className="relative transform overflow-hidden rounded-lg bg-zinc-900 shadow-xl sm:w-full sm:max-w-lg p-2" ref={ props?.ref }>
            { props.children }
          </div>
        </div>
      </div>
    </div>
  );
}
