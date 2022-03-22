import { CircularProgress as MaterialCircularProgress } from '@mui/material';
import {
  CircularProgressProps as MaterialCircularProgressProps
} from '@mui/material/CircularProgress/CircularProgress';
import classNames from 'classnames';
import React from 'react';

interface CircularProgressExtendedProps {
  visible?: boolean;
  submitting?: boolean;
}

export type CircularProgressProps = CircularProgressExtendedProps & MaterialCircularProgressProps;

export default function CircularProgress(props: CircularProgressProps): JSX.Element {
  const { submitting, visible, ...materialProps } = props;
  if (!visible) {
    return null;
  }

  materialProps.className += classNames({
    'opacity-50': submitting,
  });
  return (
    <MaterialCircularProgress
      { ...materialProps }
    />
  )
}
