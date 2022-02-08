import { Button, ButtonGroup, Divider, IconButton, List, ListItem, Typography } from '@mui/material';
import { ChevronRight } from '@mui/icons-material';
import FundingScheduleListItem from 'components/FundingSchedules/FundingScheduleListItem';
import NewFundingScheduleDialog from 'components/FundingSchedules/NewFundingScheduleDialog';
import FundingSchedule from 'models/FundingSchedule';
import { Map } from 'immutable';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import { getFundingSchedules } from 'shared/fundingSchedules/selectors/getFundingSchedules';
import NewExpenseDialog from 'components/Expenses/NewExpenseDialog';

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
        <List className="w-full pt-0" dense>
          {
            fundingSchedules.map(schedule => (
              <FundingScheduleListItem fundingScheduleId={ schedule.fundingScheduleId }/>
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
