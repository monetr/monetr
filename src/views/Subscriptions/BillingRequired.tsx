import React, { Component } from "react";
import { Redirect, Route, Switch } from "react-router-dom";
import { UpdateSubscriptionsView } from "views/Subscriptions/UpdateSubscriptionsView";


export class BillingRequired extends Component<any, any> {

  render() {
    return (
      <Switch>
        <Route path="/account/subscribe">
          <UpdateSubscriptionsView/>
        </Route>
      </Switch>
    );
  }
}
