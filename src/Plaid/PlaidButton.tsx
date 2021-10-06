import React, { Component, Fragment } from "react";
import { Button, ButtonProps, CircularProgress } from "@material-ui/core";
import { usePlaidLink } from "react-plaid-link";
import request from "shared/util/request";
import classnames from "classnames";

export interface PropTypes extends ButtonProps {
  useCache?: boolean;
  onSuccess: (token: string, metadata: object) => any;
  onExit: (event: object) => any;
  onLoad: (event: object) => any;
  onEvent: (event: object) => any;
}

interface HookedPropTypes extends PropTypes {
  token: string;
}

const HookedPlaidButton = (props: HookedPropTypes) => {
  const config = {
    token: props.token,
    onSuccess: props.onSuccess,
    onExit: props.onExit,
    onLoad: props.onLoad,
    onEvent: props.onEvent,
  };

  const { error, open } = usePlaidLink(config);

  const onClick = (event) => {
    if (props.onClick) {
      props.onClick(event);
    }

    open();
  };

  const newProps: HookedPropTypes = {
    ...props,
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
}

export default class PlaidButton extends Component<PropTypes, State> {

  state = {
    token: null,
    loading: true,
    disabled: false,
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
        console.error(error);
        this.setState({
          loading: false,
          disabled: true,
        })
      });
  }

  render() {
    const disabled = this.state.loading || this.props.disabled || this.state.disabled;
    const props: ButtonProps = {
      ...this.props,
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
        onSuccess={ this.props.onSuccess }
        onExit={ this.props.onExit }
        onEvent={ this.props.onEvent }
        onLoad={ this.props.onLoad }
        { ...props }
      />
    )
  }
}
