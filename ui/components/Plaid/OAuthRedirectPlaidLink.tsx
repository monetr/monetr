import React, { useEffect } from "react";
import { usePlaidLink } from "react-plaid-link";

export interface PropTypes {
  linkToken: string;
  onSuccess: (token: string, metadata: object) => any;
  onExit: (error: object, metadata: object) => any;
  onEvent: (event: object, metadata: object) => any;
  onLoad: (load: object, metadata: object) => any;
}

export const OAuthRedirectPlaidLink: React.FC<PropTypes> = props => {
  const { open, ready } = usePlaidLink({
    token: props.linkToken,
    receivedRedirectUri: window.location.href,
    onSuccess: props.onSuccess,
    onExit: props.onExit,
    onEvent: props.onEvent,
    onLoad: props.onLoad,
  });

  useEffect(() => {
    if (ready) {
      open();
    }
  }, [ready, open]);

  return null;
};
