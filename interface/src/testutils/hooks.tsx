import type React from 'react';
import NiceModal from '@ebay/nice-modal-react';
import { Router } from 'wouter';

import { type RenderHookResult, renderHook } from '@testing-library/react';

import MQueryClient from '@monetr/interface/components/MQueryClient';
import MSnackbarProvider from '@monetr/interface/components/MSnackbarProvider';

import { memoryLocation } from 'wouter/memory-location';

export interface HooksOptions<TProps> {
  initialRoute: string;
  initialProps?: TProps;
}

function testRenderHook<TProps, TResult>(
  callback: (props: TProps) => TResult,
  options?: HooksOptions<TProps>,
): RenderHookResult<TResult, TProps> {
  const { hook } = memoryLocation({ path: options?.initialRoute ?? '/' });
  const Wrapper: React.FC<React.PropsWithChildren> = props => {
    return (
      <Router hook={hook}>
        <MQueryClient>
          <MSnackbarProvider>
            <NiceModal.Provider>{props.children}</NiceModal.Provider>
          </MSnackbarProvider>
        </MQueryClient>
      </Router>
    );
  };

  return renderHook<TResult, TProps>(callback, {
    wrapper: Wrapper,
    initialProps: options?.initialProps,
  });
}

export default testRenderHook;
