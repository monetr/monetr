import React, { Component, Fragment } from 'react';
import { Alert, AlertTitle, Button, ButtonProps, CircularProgress, Snackbar } from '@mui/material';
import {
  usePlaidLink,
  PlaidLinkOptionsWithLinkToken,
  PlaidLinkOnEvent,
  PlaidLinkOnLoad,
  PlaidLinkOnExit, PlaidLinkOnSuccess
} from 'react-plaid-link';
import request from 'shared/util/request';
import classnames from 'classnames';

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

const HookedPlaidButton = (props: HookedPropTypes) => {
  const config: PlaidLinkOptionsWithLinkToken = {
    token: props.token,
    onSuccess: props.plaidOnSuccess,
    onExit: props.plaidOnExit,
    onLoad: props.plaidOnLoad,
    onEvent: props.plaidOnEvent,
  };

  const { error, open } = usePlaidLink(config);

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

interface State {
  token: string | null;
  disabled?: boolean;
  loading: boolean;
  error: string | null;
}

export default class PlaidButton extends Component<PropTypes, State> {

  state = {
    token: null,
    loading: true,
    disabled: false,
    error: null,
  };

  componentDidMount() {
    const url = `/plaid/link/token/new${ this.props.useCache ? '?use_cache=true' : '' }`
    request().get(url)
      .then(result => {
        this.setState({
          loading: false,
          token: result.data.linkToken,
        });
      })
      .catch(error => {
        console.error({ error });
        this.setState({
          loading: false,
          disabled: true,
          error: error?.response?.data?.error || 'Could not connect to Plaid, an unknown error occurred.'
        })
      });
  }

  renderButton = (): React.ReactNode => {
    const disabled = this.state.loading || this.props.disabled || this.state.disabled;
    // I want to extract only the button props, the easiest way to do that is to do a lift of the properties like this.
    // This unfortunately leaves a ton of variables hanging though.
    const { useCache, plaidOnSuccess, plaidOnExit, plaidOnLoad, plaidOnEvent, ...buttonProps } = this.props;
    const props: ButtonProps = {
      ...buttonProps,
      disabled: disabled,
      children: (
        <Fragment>
          { this.state.loading && <CircularProgress size="1em" thickness={ 5 } className={ classnames('mr-2', {
            'opacity-50': disabled,
          }) }/> }
          { this.props.children }
        </Fragment>
      ),
    }

    if (!this.state.token) {
      return (
        <Button { ...props } />
      );
    }

    return (
      <HookedPlaidButton
        token={ this.state.token }
        plaidOnSuccess={ this.props.plaidOnSuccess }
        plaidOnExit={ this.props.plaidOnExit }
        plaidOnEvent={ this.props.plaidOnEvent }
        plaidOnLoad={ this.props.plaidOnLoad }
        { ...props }
      />
    )
  };

  renderErrorMaybe = (): React.ReactNode => {
    const { error } = this.state;

    if (!error) {
      return null;
    }

    return (
      <Snackbar open autoHideDuration={ 10000 }>
        <Alert variant="filled" severity="error">
          <AlertTitle>Error</AlertTitle>
          { error }
        </Alert>
      </Snackbar>
    );
  };

  render() {
    return (
      <Fragment>
        { this.renderErrorMaybe() }
        { this.renderButton() }
      </Fragment>
    )
  }
}
