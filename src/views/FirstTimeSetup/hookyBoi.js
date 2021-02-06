import React from 'react';
import { usePlaidLink } from 'react-plaid-link';
import {Button} from "@material-ui/core";

export const PlaidConnectButton = props => {
  const config = {
    token: props.token,
    onSuccess: props.onSuccess
  };

  const { open, ready, error } = usePlaidLink(config);

  return (
    <Button style={{float: 'right'}} color="primary" variant="outlined" onClick={ open }>
      Connect
    </Button>
  )
}
