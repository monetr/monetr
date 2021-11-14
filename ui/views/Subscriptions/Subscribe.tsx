import request from 'shared/util/request';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import { getInitialPlan } from 'shared/bootstrap/selectors';
import { RouteComponentProps } from 'react-router-dom';
import { AppState } from 'store';

interface State {
  loading: boolean;
}

interface WithConnectionPropTypes extends RouteComponentProps {
  initialPlan: { price: number, freeTrialDays: number } | null;
}

class Subscribe extends Component<WithConnectionPropTypes, State> {

  componentDidMount() {
    this.setupStripe();
  }

  setupStripe = () => {
    const { initialPlan } = this.props;
    if (initialPlan) {
      return request().post(`/billing/create_checkout`, {
        priceId: '',
        cancelPath: '/logout',
      })
        .then(result => window.location.assign(result.data.url))
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
  (state: AppState) => ({
    initialPlan: getInitialPlan(state),
  }),
  {}
)(Subscribe);


