import React, { Component, Fragment } from "react";
import { Button, Card, CardActions, CardContent, Paper, Snackbar, Typography } from "@material-ui/core";
import BillingPlan from "data/BillingPlan";
import fetchBillingPlans from "shared/billing/actions/fetchBillingPlans";
import { getStripePublicKey } from "shared/bootstrap/selectors";
import { connect } from "react-redux";
import {
  loadStripe,
  RedirectToCheckoutOptions,
  Stripe,
  StripeCardElement,
  StripeCardElementChangeEvent,
  StripeElements
} from '@stripe/stripe-js';
import { CardElement, Elements, ElementsConsumer } from '@stripe/react-stripe-js';
import createNewSubscription from "shared/billing/actions/createNewSubscription";
import logout from "shared/authentication/actions/logout";
import { Alert, AlertTitle } from "@material-ui/lab";
import request from "shared/util/request";

enum Stage {
  ChoosePlan,
  BillingInformation,
}

interface State {
  loading: boolean;
  plans: BillingPlan[];
  stage: Stage;
  selectedPlan: BillingPlan | null;
  cardComplete: boolean;
  error: string | null;
}

interface WithConnectionPropTypes {
  stripePublicKey: string;
  logout: () => void;
}

export class UpdateSubscriptionsView extends Component<WithConnectionPropTypes, State> {

  state = {
    loading: true,
    plans: [],
    stage: Stage.ChoosePlan,
    selectedPlan: null,
    cardComplete: false,
    error: null,
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
    return this.doStripeCheckout(plan);
    // this.setState({
    //   stage: Stage.BillingInformation,
    //   selectedPlan: plan,
    // });
  };

  cancelSubscribe = () => this.setState({
    stage: Stage.ChoosePlan,
    selectedPlan: null,
  });

  doStripeCheckout = (plan: BillingPlan) => {
    return request().post(`/billing/create_checkout`, {
      priceId: plan.id,
    })
      .then(result => {
        return loadStripe(this.props.stripePublicKey).then(stripe => {
          const options: RedirectToCheckoutOptions = {
            sessionId: result.data.sessionId,
          };

          return stripe.redirectToCheckout(options);
        });
      })
      .catch(error => this.setState({
        error: error?.response?.data?.error || error,
      }));
  };

  renderCard = (details: BillingPlan) => {

    const price = `$${ (details.unitPrice / 100).toFixed(2) }`;

    return (
      <Fragment>
        { this.renderErrorMaybe() }
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
      </Fragment>
    );
  };

  renderPlanSelection = () => (
    <div className="h-full grid grid-flow-row p-10">
      <Typography className="w-full text-center mb-10 opacity-50" variant="h2">
        { this.state.loading ? 'One moment...' : 'Choose a plan that works best for you' }
      </Typography>
      <div className="w-full grid grid-flow-col gap-5">
        { this.state.plans.map(item => this.renderCard(item)) }
      </div>
      <div className="h-full w-full flex items-end justify-center opacity-50">
        <Button
          onClick={ this.props.logout }
        >
          Logout
        </Button>
      </div>
    </div>
  );

  stripeOnChange = (event: StripeCardElementChangeEvent) => {
    this.setState({
      cardComplete: event.complete,
    });
  };

  stripeOnReady = (element: StripeCardElement) => {
    console.log("ready")
  };

  stripeOnEscape = () => {

  };

  renderErrorMaybe = () => {
    const { error } = this.state;
    if (!error) {
      return null;
    }

    return (
      <Snackbar open autoHideDuration={ 10000 }>
        <Alert variant="filled" severity="error">
          <AlertTitle>Error</AlertTitle>
          { this.state.error }
        </Alert>
      </Snackbar>
    );
  };

  renderBillingHandler = (stripe: Stripe, elements: StripeElements) => {
    const { selectedPlan } = this.state;
    const price = `$${ (selectedPlan.unitPrice / 100).toFixed(2) }`;

    const billingSubscribe = () => {
      this.setState({
        loading: true,
      });
      const cardElement = elements.getElement(CardElement);
      stripe.createPaymentMethod({
        type: 'card',
        card: cardElement,
      })
        .then(result => {
          return createNewSubscription(selectedPlan.id, result.paymentMethod.id);
        })
        .then(result => {
          console.log(result);
        })
        .catch(error => this.setState({
          error: error?.response?.data?.error || error,
        }))
        .finally(() => {
          this.setState({
            loading: false,
          });
        });
    };

    return (
      <Fragment>
        { this.renderErrorMaybe() }
        <Card className="w-1/3">
          <CardContent>
            <div>
              <Typography variant="h5">Begin your subscription</Typography>
            </div>
            <div className="mt-5 mb-5">
              <Typography>Total due now: <b>{ price }</b></Typography>
            </div>
            <div className="">
              <Typography className="opacity-60">Card</Typography>
              <Paper className="p-1 pl-2">
                <CardElement
                  onChange={ this.stripeOnChange }
                  onReady={ this.stripeOnReady }
                  onEscape={ this.stripeOnEscape }
                  options={ {
                    style: {
                      base: {
                        lineHeight: '30px',
                        padding: '10px',
                      }
                    }
                  } }
                />
              </Paper>
            </div>
          </CardContent>
          <CardActions>
            <Button
              disabled={ this.state.loading }
              color="secondary"
              variant="outlined"
              className="ml-auto"
              onClick={ this.cancelSubscribe }
            >
              Cancel
            </Button>
            <Button
              disabled={ !this.state.cardComplete || this.state.loading }
              color="primary"
              variant="contained"
              onClick={ billingSubscribe }
            >
              Subscribe
            </Button>
          </CardActions>
        </Card>
      </Fragment>
    )
  };

  renderBillingInfo = () => {
    const stripePromise = loadStripe(this.props.stripePublicKey);

    return (
      <Elements stripe={ stripePromise }>
        <ElementsConsumer>
          { ({ stripe, elements }) => (
            this.renderBillingHandler(stripe, elements)
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
  {
    logout,
  },
)(UpdateSubscriptionsView);
