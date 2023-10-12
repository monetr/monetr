import React from 'react';
import { QueryClientProvider } from '@tanstack/react-query';

import { ReactElement } from './types';

import queryClient from 'client';


export interface MQueryClientProps {
  children: ReactElement;
}

export default function MQueryClient(props: MQueryClientProps): JSX.Element {
  return (
    <QueryClientProvider client={ queryClient }>
      {props.children}
    </QueryClientProvider>
  );
}
