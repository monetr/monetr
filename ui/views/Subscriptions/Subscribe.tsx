import request from 'shared/util/request';
import React, { Component } from 'react';
import { loadStripe } from '@stripe/stripe-js';
import { RedirectToCheckoutOptions } from '@stripe/stripe-js/types/stripe-js/checkout';
import { connect } from 'react-redux';
import { getInitialPlan, getStripePublicKey } from 'shared/bootstrap/selectors';
import { RouteComponentProps } from 'react-router-dom';

interface State {
  loading: boolean;
}

interface WithConnectionPropTypes extends RouteComponentProps {
  initialPlan: { price: number, freeTrialDays: number } | null;
  stripePublicKey: string | null;
}

class Subscribe extends Component<WithConnectionPropTypes, State> {

  componentDidMount() {
    this.setupStripe();
  }

  setupStripe = () => {
    const { initialPlan, stripePublicKey } = this.props;
    if (initialPlan) {
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

    this.setState({
      loading: false,
    });

    return Promise.resolve();
  };

  render() {
    return null;
  }
}

export default connect(
  state => ({
    initialPlan: getInitialPlan(state),
    stripePublicKey: getStripePublicKey(state),
  }),
  {}
)(Subscribe);


