import React, { Component } from "react";
import { Button, Card, CardContent, Typography } from "@material-ui/core";
import BillingPlan from "data/BillingPlan";
import fetchBillingPlans from "shared/billing/actions/fetchBillingPlans";

interface State {
  loading: boolean;
  plans: BillingPlan[];
}

export class UpdateSubscriptionsView extends Component<any, State> {

  state = {
    loading: true,
    plans: [],
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
    console.log(plan);
  };

  renderCard = (details: BillingPlan) => {

    const price = `$${ (details.unitPrice / 100).toFixed(2) }`;

    return (
      <div key={ details.id } className="flex justify-center items-center">
        <Card className="transition transform hover:shadow-2xl hover:scale-105 w-64 h-72">
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

  render() {
    return (
      <div className="w-full h-full flex justify-center items-center">
        <div className="grid grid-flow-row">
          <Typography className="w-full text-center mb-10 opacity-50" variant="h2">
            { this.state.loading ? 'One moment...' : 'Choose a plan that works best for you' }
          </Typography>
          <div className="w-full grid grid-flow-col gap-10">
            { this.state.plans.map(item => this.renderCard(item)) }
          </div>
        </div>
      </div>
    );
  }
}
