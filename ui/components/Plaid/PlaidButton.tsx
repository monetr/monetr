import { useSnackbar } from 'notistack';
import React, { Fragment, useEffect, useState } from 'react';
import { Button, ButtonProps, CircularProgress } from '@mui/material';
import {
  usePlaidLink,
  PlaidLinkOptionsWithLinkToken,
  PlaidLinkOnEvent,
  PlaidLinkOnLoad,
  PlaidLinkOnExit, PlaidLinkOnSuccess
} from 'react-plaid-link';
import request from 'shared/util/request';
import classnames from 'classnames';
import * as Sentry from '@sentry/react';

interface BasePropTypes {
  useCache?: boolean;
  plaidOnSuccess: PlaidLinkOnSuccess;
  plaidOnExit?: PlaidLinkOnExit;
  plaidOnLoad?: PlaidLinkOnLoad;
  plaidOnEvent?: PlaidLinkOnEvent;
}

type PropTypes = BasePropTypes & ButtonProps;

interface HookedPropTypes extends PropTypes {
  token: string;
}

interface State {
  token: string | null;
  disabled?: boolean;
  loading: boolean;
  error: string | null;
}

const PlaidButton = (props: PropTypes): JSX.Element => {
  const { enqueueSnackbar } = useSnackbar();
  const [state, setState] = useState<Partial<State>>({});

  useEffect(() => {
    const url = `/plaid/link/token/new${ props.useCache ? '?use_cache=true' : '' }`
    request().get(url)
      .then(result => setState({
        loading: false,
        token: result.data.linkToken,
      }))
      .catch(error => {
        console.error({ error });
        setState({
          loading: false,
          disabled: true,
        });
        enqueueSnackbar(error?.response?.data?.error || 'Could not connect to Plaid, an unknown error occurred.', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      });
  }, []);

  if (!state.token) {
    const disabled = state.loading || props.disabled || state.disabled;
    // I want to extract only the button props, the easiest way to do that is to do a lift of the properties like this.
    // This unfortunately leaves a ton of variables hanging though.
    const { useCache, plaidOnSuccess, plaidOnExit, plaidOnLoad, plaidOnEvent, ...buttonProps } = props;
    const newProps: ButtonProps = {
      ...buttonProps,
      disabled: disabled,
      children: (
        <Fragment>
          { state.loading && <CircularProgress size="1em" thickness={ 5 } className={ classnames('mr-2', {
            'opacity-50': disabled,
          }) }/> }
          { props.children }
        </Fragment>
      ),
    }

    return (
      <Button { ...newProps } />
    );
  }

  return (
    <HookedPlaidButton
      token={ state.token }
      plaidOnSuccess={ props.plaidOnSuccess }
      plaidOnExit={ props.plaidOnExit }
      plaidOnEvent={ props.plaidOnEvent }
      plaidOnLoad={ props.plaidOnLoad }
      { ...props }
    />
  )

};

const HookedPlaidButton = (props: HookedPropTypes) => {
  const { enqueueSnackbar } = useSnackbar();
  const config: PlaidLinkOptionsWithLinkToken = {
    token: props.token,
    onSuccess: props.plaidOnSuccess,
    onExit: props.plaidOnExit,
    onLoad: props.plaidOnLoad,
    onEvent: props.plaidOnEvent,
  };

  const { error, open } = usePlaidLink(config);

  useEffect(() => {
    if (error) {
      Sentry.captureException(error);
      enqueueSnackbar('Failed to setup Plaid link.', {
        variant: 'error',
        disableWindowBlurListener: true,
      });
    }
  }, [error]);

  const onClick = (event) => {
    if (props.onClick) {
      props.onClick(event);
    }

    open();
  };

  // I want to extract only the button props, the easiest way to do that is to do a lift of the properties like this.
  // This unfortunately leaves a ton of variables hanging though.
  const { useCache, plaidOnSuccess, plaidOnExit, plaidOnLoad, plaidOnEvent, token, ...buttonProps } = props;

  const newProps = {
    ...buttonProps,
    onClick,
  };

  return (
    <Button { ...newProps } />
  );
};

export default PlaidButton;