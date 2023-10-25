import React from 'react';
import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import { SnackbarProvider, VariantType } from 'notistack';

import { ReactElement } from './types';

const snackbarIcons: Partial<Record<VariantType, React.ReactNode>> = {
  error: <ErrorIcon className="mr-2.5" />,
  success: <DoneIcon className="mr-2.5" />,
  warning: <WarningIcon className="mr-2.5" />,
  info: <InfoIcon className="mr-2.5" />,
};

export interface MSnackbarProviderProps {
  children: ReactElement
}

export default function MSnackbarProvider(props: MSnackbarProviderProps): JSX.Element {
  return (
    <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
      { props.children }
    </SnackbarProvider>
  );
}
