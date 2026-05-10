import type React from 'react';
import { Check, CircleAlert, Info, TriangleAlert } from 'lucide-react';
import { SnackbarProvider, type VariantType } from 'notistack';

import styles from './MSnackbarProvider.module.scss';

const snackbarIcons: Partial<Record<VariantType, React.ReactNode>> = {
  error: <CircleAlert className={styles.icon} />,
  success: <Check className={styles.icon} />,
  warning: <TriangleAlert className={styles.icon} />,
  info: <Info className={styles.icon} />,
};

export interface MSnackbarProviderProps {
  children: React.ReactNode;
}

export default function MSnackbarProvider(props: MSnackbarProviderProps): JSX.Element {
  return (
    <SnackbarProvider iconVariant={snackbarIcons} maxSnack={5}>
      {props.children}
    </SnackbarProvider>
  );
}
