import React, { Component } from "react";
import { Button, Card, CardContent, Typography } from "@material-ui/core";
import BillingPlan from "data/BillingPlan";
import fetchBillingPlans from "shared/billing/actions/fetchBillingPlans";
import { getStripePublicKey } from "shared/bootstrap/selectors";
import { connect } from "react-redux";
import { loadStripe } from '@stripe/stripe-js';
import { CardElement, Elements, ElementsConsumer } from '@stripe/react-stripe-js';

enum Stage {
  ChoosePlan,
  BillingInformation,
}

interface State {
  loading: boolean;
  plans: BillingPlan[];
  stage: Stage;
}

interface WithConnectionPropTypes {
  stripePublicKey: string;
}

export class UpdateSubscriptionsView extends Component<WithConnectionPropTypes, State> {

  state = {
    loading: true,
    plans: [],
    stage: Stage.ChoosePlan,
  };

  componentDidMount() {
    fetchBillingPlans().then(result => {
      this.setState({
        plans: result,
        loading: false,
      });
    });
  }

  selectPlan = (plan: BillingPlan) => () => {
    this.setState({
      stage: Stage.BillingInformation,
    });
  };


  renderCard = (details: BillingPlan) => {

    const price = `$${ (details.unitPrice / 100).toFixed(2) }`;

    return (
      <div key={ details.id } className="flex justify-center items-center">
        <Card elevation={ 4 } className="smooth-animation transition transform hover:shadow-2xl w-64 h-72">
          <CardContent className="h-full">
            <div className="grid grid-flow-row h-full">
              <div>
                <div className="grid grid-flow-col">
                  <Typography variant='h5'>
                    <b>{ details.name }</b>
                  </Typography>
                  <div className="flex items-center justify-end">
                    <Typography variant="body1">
                      <b>{ price }</b> / { details.interval }
                    </Typography>
                  </div>
                </div>
                <Typography variant="body2">
                  { details.description }
                </Typography>
              </div>
              <div className="flex items-end grid grid-flow-row">
                { (details.freeTrialDays > 0) &&
                <Typography
                  variant="h6"
                  className="w-full text-center"
                >
                  { details.freeTrialDays } day free trial
                </Typography>
                }
                <Button
                  className="w-full h-12"
                  color="primary"
                  variant="contained"
                  onClick={ this.selectPlan(details) }
                >
                  <Typography
                    variant="h6"
                  >
                    Choose
                  </Typography>
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  };

  renderPlanSelection = () => (
    <div className="grid grid-flow-row">
      <Typography className="w-full text-center mb-10 opacity-50" variant="h2">
        { this.state.loading ? 'One moment...' : 'Choose a plan that works best for you' }
      </Typography>
      <div className="w-full grid grid-flow-col gap-5">
        { this.state.plans.map(item => this.renderCard(item)) }
      </div>
    </div>
  );

  renderBillingInfo = () => {
    const stripePromise = loadStripe(this.props.stripePublicKey);

    return (
      <Elements stripe={ stripePromise }>
        <ElementsConsumer>
          { ({ stripe, elements }) => (
            <div className="w-96 h-64">
              <CardElement className="h-12"/>
            </div>
          ) }
        </ElementsConsumer>
      </Elements>
    );
  };

  renderContents = () => {
    switch (this.state.stage) {
      case Stage.ChoosePlan:
        return this.renderPlanSelection();
      case Stage.BillingInformation:
        return this.renderBillingInfo();
    }
  };

  render() {
    return (
      <div className="w-full h-full flex justify-center items-center">
        { this.renderContents() }
      </div>
    );
  }
}

export default connect(
  state => ({
    stripePublicKey: getStripePublicKey(state),
  }),
  {},
)(UpdateSubscriptionsView);
