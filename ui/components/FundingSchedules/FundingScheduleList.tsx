import { Button, ButtonGroup, Divider, IconButton, List, ListItem, Typography } from "@material-ui/core";
import { ChevronRight } from '@material-ui/icons';
import NewFundingScheduleDialog from "components/FundingSchedules/NewFundingScheduleDialog";
import FundingSchedule from "models/FundingSchedule";
import { Map } from 'immutable';
import React, { Component } from "react";
import { connect } from "react-redux";
import { getFundingSchedules } from "shared/fundingSchedules/selectors/getFundingSchedules";
import NewExpenseDialog from "components/Expenses/NewExpenseDialog";

interface WithConnectionPropTypes {
  fundingSchedules: Map<number, FundingSchedule>;
}

interface State {
  newFundingScheduleDialogOpen: boolean;
  newExpenseDialogOpen: boolean;
}

export class FundingScheduleList extends Component<WithConnectionPropTypes, State> {

  state = {
    newFundingScheduleDialogOpen: false,
    newExpenseDialogOpen: false,
  };

  openNewExpenseDialog = () => {
    return this.setState({
      newExpenseDialogOpen: true
    });
  };

  closeNewExpenseDialog = () => {
    return this.setState({
      newExpenseDialogOpen: false
    });
  };

  openNewFundingScheduleDialog = () => {
    return this.setState({
      newFundingScheduleDialogOpen: true,
    });
  };

  closeFundingScheduleDialog = () => {
    return this.setState({
      newFundingScheduleDialogOpen: false,
    });
  };

  render() {
    const { fundingSchedules } = this.props;
    return (
      <div className="w-full funding-schedule-list">
        { this.state.newFundingScheduleDialogOpen &&
        <NewFundingScheduleDialog
          onClose={ this.closeFundingScheduleDialog }
          isOpen={ this.state.newFundingScheduleDialogOpen }
        />
        }
        { this.state.newExpenseDialogOpen &&
        <NewExpenseDialog
          onClose={ this.closeNewExpenseDialog }
          isOpen={ this.state.newExpenseDialogOpen }
        />
        }
        <div className="w-full p-5">
          <ButtonGroup color="primary" className="w-full">
            <Button variant="outlined" className="w-full" color="primary" onClick={ this.openNewFundingScheduleDialog }>
              New Funding Schedule
            </Button>
            <Button variant="outlined" className="w-full" color="primary" onClick={ this.openNewExpenseDialog }>
              New Expense
            </Button>
          </ButtonGroup>
        </div>
        <Divider/>
        <List className="w-full">
          {
            fundingSchedules.map(schedule => (
              <ListItem key={ schedule.fundingScheduleId } button>
                <div className="grid grid-cols-4 grid-rows-2 grid-flow-col gap-1 w-full">
                  <div className="col-span-2">
                    <Typography>{ schedule.name }</Typography>
                  </div>
                  <div className="col-span-3 opacity-50">
                    <Typography variant="body2">{ schedule.description }</Typography>
                  </div>
                  <div className="col-span-1 flex justify-end">
                    <Typography variant="subtitle2"
                                color="primary">{ schedule.nextOccurrence.format('MMM Do') }</Typography>
                  </div>
                  <div className="row-span-2 col-span-1 flex justify-end">
                    <IconButton>
                      <ChevronRight/>
                    </IconButton>
                  </div>
                </div>
              </ListItem>
            )).valueSeq().toArray()
          }
        </List>
      </div>
    )
  }
}

export default connect(
  state => ({
    fundingSchedules: getFundingSchedules(state),
  }),
  {}
)(FundingScheduleList);
