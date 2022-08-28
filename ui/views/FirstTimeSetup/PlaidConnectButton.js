import React from 'react';
import { usePlaidLink } from 'react-plaid-link';
import { Button } from '@mui/material';
import * as Sentry from '@sentry/react';

export const PlaidConnectButton = props => {
  const config = {
    token: props.token,
    onSuccess: props.onSuccess,
    onExit: props.onExit,
    onLoad: props.onLoad,
    onEvent: props.onEvent,
  };


  const { error, open } = usePlaidLink(config);
  if (error) {
    console.warn({
      error,
    });
    Sentry.captureException(error);
  }

  const onClick = () => {

    if (props.onClick) {
      props.onClick();
    }

    open();
  };

  return (
    <Button
      disabled={ props.disabled }
      style={ { float: 'right' } }
      color="primary"
      variant="outlined"
      onClick={ onClick }>
      Connect
    </Button>
  );
};
