import React from "react";
import { Redirect, Route, Switch } from "react-router-dom";

import Logout from "views/Authentication/Logout";
import AfterCheckout from "views/Subscriptions/AfterCheckout";
import Subscribe from "views/Subscriptions/Subscribe";

export default function BillingRequiredRouter(): React.ReactNode {
  return (
    <Switch>
      <Route path="/logout" exact component={ Logout }/>
      <Route path="/account/subscribe" exact component={ Subscribe }/>
      <Route path="/account/subscribe/after" exact component={ AfterCheckout }/>
      <Route>
        <Redirect to="/account/subscribe"/>
      </Route>
    </Switch>
  );
}
