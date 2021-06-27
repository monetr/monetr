import React, { Component } from "react";
import { connect } from "react-redux";
import { RouteComponentProps, withRouter } from "react-router-dom";

import Logo from 'assets';
import request from "shared/util/request";

import { CircularProgress, Typography } from "@material-ui/core";

interface State {
  loading: boolean;
  longPollAttempts: number;
}

class AfterCheckout extends Component<RouteComponentProps, State> {

  state = {
    loading: true,
    longPollAttempts: 0,
  };

  componentDidMount() {
    // This component is meant as a loading state for after the user has come back from the stripe checkout session.
    // As soon as the component mounts start the polling and wait for the subscription.
    this.longPollSetup();
  }

  longPollSetup = () => {
    this.setState(prevState => ({
      longPollAttempts: prevState.longPollAttempts + 1,
    }));

    const { longPollAttempts } = this.state;
    if (longPollAttempts > 6) {
      return Promise.resolve();
    }

    return request().get(`/billing/wait`)
      .then(() => {
        this.props.history.push('/setup');
        return Promise.resolve();
      })
      .catch(error => {
        if (error.response.status === 408) {
          return this.longPollSetup();
        }

        console.warn(error);

        throw error;
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
  {}
)(withRouter(AfterCheckout));
