import React, { Component } from "react";
import { Route, Switch, withRouter } from "react-router-dom";
import { connect } from "react-redux";

import Logout from "views/Authentication/Logout";
import { getInitialPlan, getStripePublicKey } from "shared/bootstrap/selectors";
import AfterCheckout from "views/Subscriptions/AfterCheckout";
import Subscribe from "views/Subscriptions/Subscribe";

interface State {
}


class BillingRequired extends Component<any, State> {

  render() {
    return (
      <Switch>
        <Route path="/logout" component={ Logout }/>
        <Route path="/account/subscribe" exact component={ Subscribe } />
        <Route path="/account/subscribe/after" exact component={ AfterCheckout }/>
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
