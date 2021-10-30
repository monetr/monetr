import React, { Component } from "react";
import { connect } from "react-redux";
import { RouteComponentProps, withRouter } from "react-router-dom";

import Logo from 'assets';
import request from "shared/util/request";

import { CircularProgress, Typography } from "@mui/material";
import activateSubscription from "shared/authentication/actions/activateSubscription";

interface State {
  loading: boolean;
}

interface WithConnectionProps extends RouteComponentProps {
  activateSubscription: () => void;
}

class AfterCheckout extends Component<WithConnectionProps, State> {

  state = {
    loading: true,
  };

  componentDidMount() {
    // This component is meant as a loading state for after the user has come back from the stripe checkout session.
    // As soon as the component mounts start the polling and wait for the subscription.
    this.setupFromCheckout();
  }

  setupFromCheckout = () => {
    const params = new URLSearchParams(this.props.location.search);
    const checkoutSessionId  = params.get('session');
    return request().get(`/billing/checkout/${checkoutSessionId}`)
      .then(result => {
        const { data } = result;
        if (data.isActive) {
          this.props.activateSubscription();
          this.props.history.push('/');
          return;
        }

        alert('subscription is not active');
      })
      .catch(error => {
        console.log(error);
      });
  };

  render() {
    return (
      <div className="flex items-center justify-center w-full h-full max-h-full">
        <div className="w-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
          <div className="flex justify-center w-full mb-5">
            <img src={ Logo } className="w-1/3"/>
          </div>
          <div className="w-full pt-2.5 pb-2.5">
            <Typography
              variant="h5"
              className="w-full text-center"
            >
              Getting your account setup...
            </Typography>
          </div>
          <div className="w-full pt-2.5 pb-2.5 flex justify-center">
            <CircularProgress />
          </div>
        </div>
      </div>
    )
  }
}

export default connect(
  _ => ({}),
  {
    activateSubscription,
  }
)(withRouter(AfterCheckout));
