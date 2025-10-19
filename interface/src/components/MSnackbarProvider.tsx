import type React from 'react';
import { Check, CircleAlert, Info, TriangleAlert } from 'lucide-react';
import { SnackbarProvider, type VariantType } from 'notistack';

const snackbarIcons: Partial<Record<VariantType, React.ReactNode>> = {
  error: <CircleAlert className='mr-2.5' />,
  success: <Check className='mr-2.5' />,
  warning: <TriangleAlert className='mr-2.5' />,
  info: <Info className='mr-2.5' />,
};

export interface MSnackbarProviderProps {
  children: React.ReactNode;
}

export default function MSnackbarProvider(props: MSnackbarProviderProps): JSX.Element {
  return (
    <SnackbarProvider maxSnack={5} iconVariant={snackbarIcons}>
      {props.children}
    </SnackbarProvider>
  );
}
