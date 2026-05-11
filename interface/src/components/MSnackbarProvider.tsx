import type React from 'react';

import { SnackbarProvider } from '@monetr/notify';

export interface MSnackbarProviderProps {
  children: React.ReactNode;
}

export default function MSnackbarProvider(props: MSnackbarProviderProps): JSX.Element {
  return <SnackbarProvider maxSnack={5}>{props.children}</SnackbarProvider>;
}
