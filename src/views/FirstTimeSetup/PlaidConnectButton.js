import { Button } from "@material-ui/core";
import React from 'react';
import { usePlaidLink } from 'react-plaid-link';

export const PlaidConnectButton = props => {
  const config = {
    token: props.token,
    onSuccess: props.onSuccess,
    onExist: props.onExit,
    onLoad: props.onLoad,
    onEvent: props.onEvent,
  };

  const { open } = usePlaidLink(config);

  return (
    <Button
      style={ { float: 'right' } }
      color="primary"
      variant="outlined"
      onClick={ open }>
      Connect
    </Button>
  )
}
