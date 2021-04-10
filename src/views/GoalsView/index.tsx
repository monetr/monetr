import { Button, Card, Divider, List, Typography } from '@material-ui/core';
import GoalDetails from 'components/Goals/GoalDetails';
import GoalRow from 'components/Goals/GoalRow';
import NewGoalDialog from 'components/Goals/NewGoalDialog';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { getGoalIds } from 'shared/spending/selectors/getGoalIds';

import './styles/GoalsView.scss';

interface WithConnectionProps {
  goalIds: number[];
}

interface State {
  newGoalDialogOpen: boolean;
}

export class GoalsView extends Component<WithConnectionProps, State> {

  state = {
    newGoalDialogOpen: false,
  };

  openNewGoalDialog = () => this.setState({
    newGoalDialogOpen: true,
  });

  closeNewGoalDialog = () => this.setState({
    newGoalDialogOpen: false,
  });

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
    const { newGoalDialogOpen } = this.state;

    if (goalIds.length === 0) {
      return (
        <Fragment>
          { newGoalDialogOpen && <NewGoalDialog onClose={ this.closeNewGoalDialog } isOpen={ newGoalDialogOpen }/> }

          <div className="minus-nav">
            <div className="flex flex-col h-full p-10 max-h-full">
              <div className="grid grid-cols-3 gap-4 flex-grow">
                <div className="col-span-3">
                  <Card elevation={ 4 } className="w-full goals-list ">
                    <div className="h-full flex justify-center items-center">
                      <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-2">
                        <Typography
                          className="opacity-50"
                          variant="h3"
                        >
                          You don't have any goals yet...
                        </Typography>
                        <Button
                          onClick={ this.openNewGoalDialog }
                          color="primary"
                        >
                          <Typography
                            variant="h6"
                          >
                            Create A Goal
                          </Typography>
                        </Button>
                      </div>
                    </div>
                  </Card>
                </div>
              </div>
            </div>
          </div>
        </Fragment>
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
                <GoalDetails />
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
