import type React from 'react';
import NiceModal from '@ebay/nice-modal-react';
import type { AxiosInstance } from 'axios';
import { type Location, MemoryRouter } from 'react-router-dom';

import { type Queries, type queries, type RenderOptions, type RenderResult, render } from '@testing-library/react';

import MQueryClient from '@monetr/interface/components/MQueryClient';
import MSnackbarProvider from '@monetr/interface/components/MSnackbarProvider';
import { TooltipProvider } from '@monetr/interface/components/Tooltip';

export interface Options<Q extends Queries = typeof queries, Container extends Element | DocumentFragment = HTMLElement>
  extends RenderOptions<Q, Container> {
  initialRoute: string | Partial<Location>;
  client?: AxiosInstance;
}

function testRenderer<Q extends Queries = typeof queries, Container extends Element | DocumentFragment = HTMLElement>(
  ui: React.ReactElement,
  options?: Options<Q, Container>,
): RenderResult<Q, Container> {
  const Wrapper = (props: React.PropsWithChildren<any>) => {
    return (
      <MemoryRouter initialEntries={[options.initialRoute]}>
        <MQueryClient client={options.client}>
          <MSnackbarProvider>
            <TooltipProvider>
              <NiceModal.Provider>{props.children}</NiceModal.Provider>
            </TooltipProvider>
          </MSnackbarProvider>
        </MQueryClient>
      </MemoryRouter>
    );
  };

  return render<Q, Container>(ui, { wrapper: Wrapper, ...options });
}

export default testRenderer;
