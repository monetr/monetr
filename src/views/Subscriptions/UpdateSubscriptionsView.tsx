import React, { Component } from "react";
import { Button, Card, CardContent, Typography } from "@material-ui/core";


interface SubscriptionDetails {
  id: string;
  name: string;
  description: string;
  unitPrice: number;
  interval: string;
  intervalCount: number;
  trialDays: number;
}

export class UpdateSubscriptionsView extends Component<any, any> {

  renderCard = (details: SubscriptionDetails) => {

    return (
      <div key={ details.id } className="flex justify-center items-center">
        <Card className="transition transform hover:shadow-2xl hover:scale-105 w-64 h-64">
          <CardContent className="h-full">
            <Typography variant='h4'>
              { details.name }
            </Typography>
            <Typography variant="body1">
              { details.description }
            </Typography>
            <Typography variant="body1">
              { details.unitPrice } / { details.interval }
            </Typography>
            { (details.trialDays > 0) &&
              <Typography>
                { details.trialDays } day free trial
              </Typography>
            }
            <Button className="absolute bottom-2.5" color="primary">
              Choose
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  };

  render() {
    const subscriptions: SubscriptionDetails[] = [
      {
        id: 'manual',
        name: 'Manual',
        description: 'Manually manage transactions and balances.',
        unitPrice: 199,
        interval: 'month',
        intervalCount: 1,
        trialDays: 30,
      },
      {
        id: 'linked',
        name: 'Linked',
        description: 'Link your bank accounts for automatically retrieving transactions and balances.',
        unitPrice: 499,
        interval: 'month',
        intervalCount: 1,
        trialDays: 0,
      }
    ];


    return (
      <div className="w-full h-full flex justify-center items-center">
        <div className="grid grid-flow-row">
          <Typography className="w-full text-center mb-10" variant="h2">
            Please pick a plan to continue...
          </Typography>
          <div className="w-full grid grid-flow-col gap-10">
            { subscriptions.map(item => this.renderCard(item)) }
          </div>
        </div>
      </div>
    )
  }
}
