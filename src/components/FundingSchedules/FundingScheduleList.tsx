import { Button, Divider, IconButton, List, ListItem, Typography } from "@material-ui/core";
import { ChevronRight } from '@material-ui/icons';
import NewFundingScheduleDialog from "components/FundingSchedules/NewFundingScheduleDialog";
import FundingSchedule from "data/FundingSchedule";
import { Map } from 'immutable';
import React, { Component } from "react";
import { connect } from "react-redux";
import { getFundingSchedules } from "shared/fundingSchedules/selectors/getFundingSchedules";

export interface PropTypes {
  onHide: { (): void }
}

interface WithConnectionPropTypes extends PropTypes {
  fundingSchedules: Map<number, FundingSchedule>;
}


interface State {
  newFundingScheduleDialogOpen: boolean;
}

export class FundingScheduleList extends Component<WithConnectionPropTypes, State> {

  state = {
    newFundingScheduleDialogOpen: false,
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
    const { fundingSchedules, onHide } = this.props;
    return (
      <div className="w-full funding-schedule-list">
        <NewFundingScheduleDialog
          onClose={ this.closeFundingScheduleDialog }
          isOpen={ this.state.newFundingScheduleDialogOpen }
        />
        <div className="w-full p-5 grid grid-cols-3 gap-2 flex-grow">
          <div className="col-span-1">
            <Button onClick={ onHide }>
              Back
            </Button>
          </div>
          <div className="col-span-2 flex justify-end w-full">
            <Button variant="outlined" color="primary" onClick={ this.openNewFundingScheduleDialog }>
              New Funding Schedule
            </Button>
          </div>
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
