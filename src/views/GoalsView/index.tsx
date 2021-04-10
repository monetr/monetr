import { Button, Card, Divider, List, Typography } from '@material-ui/core';
import GoalRow from 'components/Goals/GoalRow';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { getGoalIds } from 'shared/spending/selectors/getGoalIds';

import './styles/GoalsView.scss';

interface WithConnectionProps {
  goalIds: number[];
}

export class GoalsView extends Component<WithConnectionProps, any> {

  renderGoalList = () => {
    const { goalIds } = this.props;

    return (
      <Card elevation={ 4 } className="w-full goals-list">
        <List disablePadding className="w-full">
          {
            goalIds.map(item => (
              <Fragment>
                <GoalRow goalId={ item } key={ item }/>
                <Divider/>
              </Fragment>
            ))
          }
        </List>
      </Card>
    )
  };

  render() {
    const { goalIds } = this.props;

    if (goalIds.length === 0) {
      return (
        <div className="minus-nav">
          <div className="flex flex-col h-full p-10 max-h-full">
            <div className="grid grid-cols-3 gap-4 flex-grow">
              <div className="col-span-3">
                <Card elevation={ 4 } className="w-full goals-list flex justify-center items-center">
                  <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-2">
                    <Typography
                      className="opacity-50"
                      variant="h3"
                    >
                      You don't have any goals yet...
                    </Typography>
                    <Button
                      color="primary"
                    >
                      <Typography
                        variant="h6"
                      >
                        Create A Goal
                      </Typography>
                    </Button>
                  </div>
                </Card>
              </div>
            </div>
          </div>
        </div>
      )
    }

    return (
      <div className="minus-nav">
        <div className="flex flex-col h-full p-10 max-h-full">
          <div className="grid grid-cols-3 gap-4 flex-grow">
            <div className="col-span-2">
              { this.renderGoalList() }
            </div>
            <div>
              <Card elevation={ 4 } className="w-full goals-list">

              </Card>
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default connect(
  (state) => ({
    goalIds: getGoalIds(state),
  }),
  {}
)(GoalsView);
