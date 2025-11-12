import type React from 'react';
import NiceModal from '@ebay/nice-modal-react';
import { MemoryRouter } from 'react-router-dom';

import { type RenderHookResult, renderHook, type WrapperComponent } from '@testing-library/react-hooks';

import MQueryClient from '@monetr/interface/components/MQueryClient';
import MSnackbarProvider from '@monetr/interface/components/MSnackbarProvider';

export interface HooksOptions<TProps> {
  initialRoute: string;
  initialProps?: TProps;
}

function testRenderHook<TProps, TResult>(
  callback: (props: TProps) => TResult,
  options?: HooksOptions<TProps>,
): RenderHookResult<TProps, TResult> {
  const Wrapper: WrapperComponent<TProps> = (props: React.PropsWithChildren<unknown>) => {
    return (
      <MemoryRouter
        future={{ v7_startTransition: false, v7_relativeSplatPath: false }}
        initialEntries={[options.initialRoute]}
      >
        <MQueryClient>
          <MSnackbarProvider>
            <NiceModal.Provider>{props.children}</NiceModal.Provider>
          </MSnackbarProvider>
        </MQueryClient>
      </MemoryRouter>
    );
  };

  return renderHook<TProps, TResult>(callback, {
    wrapper: Wrapper,
    initialProps: options?.initialProps,
  });
}

export default testRenderHook;
