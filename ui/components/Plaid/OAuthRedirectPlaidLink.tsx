import React, { useEffect } from 'react';
import { PlaidLinkOnEvent, PlaidLinkOnExit, PlaidLinkOnLoad, PlaidLinkOnSuccess, usePlaidLink } from 'react-plaid-link';

export interface PropTypes {
  linkToken: string;
  plaidOnSuccess: PlaidLinkOnSuccess;
  plaidOnExit?: PlaidLinkOnExit;
  plaidOnLoad?: PlaidLinkOnLoad;
  plaidOnEvent?: PlaidLinkOnEvent;
}

export const OAuthRedirectPlaidLink: React.FC<PropTypes> = props => {
  const { open, ready } = usePlaidLink({
    token: props.linkToken,
    receivedRedirectUri: window.location.href,
    onSuccess: props.plaidOnSuccess,
    onExit: props.plaidOnExit,
    onEvent: props.plaidOnEvent,
    onLoad: props.plaidOnLoad,
  });

  useEffect(() => {
    if (ready) {
      open();
    }
  }, [ready, open]);

  return null;
};
