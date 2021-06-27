import React, { Component } from "react";
import { Route, RouteComponentProps, Switch, withRouter } from "react-router-dom";
import { connect } from "react-redux";

import Logout from "views/Authentication/Logout";
import UpdateSubscriptionsView from "views/Subscriptions/UpdateSubscriptionsView";
import request from "shared/util/request";
import { getInitialPlan, getStripePublicKey } from "shared/bootstrap/selectors";

import { CircularProgress } from "@material-ui/core";
import { loadStripe, RedirectToCheckoutOptions } from "@stripe/stripe-js";

interface State {
  loading: boolean;
}

interface WithConnectionPropTypes extends RouteComponentProps {
  initialPlan: { price: number, freeTrialDays: number } | null;
  stripePublicKey: string | null;
}

class BillingRequired extends Component<WithConnectionPropTypes, State> {

  state = {
    loading: true,
  };

  componentDidMount() {
    this.setupStripe();
  }

  setupStripe = () => {
    const { initialPlan, stripePublicKey, history } = this.props;
    if (initialPlan) {
      switch (history.location.pathname) {
        case '/logout':
          break;
        default:
          return request().post(`/billing/create_checkout`, {
            priceId: '',
            cancelPath: '/logout',
          })
            .then(result => {
              return loadStripe(stripePublicKey).then(stripe => {
                const options: RedirectToCheckoutOptions = {
                  sessionId: result.data.sessionId,
                };

                return stripe.redirectToCheckout(options);
              });
            })
            .catch(error => alert(error));
      }
    }

    this.setState({
      loading: false,
    });

    return Promise.resolve();
  };

  render() {
    const { loading } = this.state;

    if (loading) {
      return <CircularProgress/>
    }

    return (
      <Switch>
        <Route path="/logout">
          <Logout/>
        </Route>
        <Route path="/account/subscribe">
          <UpdateSubscriptionsView/>
        </Route>
      </Switch>
    );
  }
}

export default connect(
  state => ({
    initialPlan: getInitialPlan(state),
    stripePublicKey: getStripePublicKey(state),
  }),
  {}
)(withRouter(BillingRequired));
