import { Button, Checkbox, Divider, List, ListItem, ListItemIcon, Typography } from '@material-ui/core';
import NewFundingScheduleDialog from 'components/FundingSchedules/NewFundingScheduleDialog';
import FundingSchedule from 'data/FundingSchedule';
import { Map } from 'immutable';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import { getFundingSchedules } from 'shared/fundingSchedules/selectors/getFundingSchedules';


export interface PropTypes {
  onChange: { (fundingSchedule: FundingSchedule): void }
  disabled?: boolean;
}

interface WithConnectionPropTypes extends PropTypes {
  fundingSchedules: Map<number, FundingSchedule>;
}

interface State {
  newFundingScheduleDialogOpen: boolean;
  selectedFundingSchedule?: number;
}

export class FundingScheduleSelectionList extends Component<WithConnectionPropTypes, State> {

  state = {
    newFundingScheduleDialogOpen: false,
    selectedFundingSchedule: null,
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

  selectItem = (fundingScheduleId: number) => () => {
    const { onChange, fundingSchedules } = this.props;

    return this.setState({
      selectedFundingSchedule: fundingScheduleId,
    }, () => onChange(fundingSchedules.get(fundingScheduleId)));
  };

  render() {
    const { fundingSchedules, disabled } = this.props;
    const { selectedFundingSchedule } = this.state;

    return (
      <div className="w-full funding-schedule-selection-list">
        <NewFundingScheduleDialog
          onClose={ this.closeFundingScheduleDialog }
          isOpen={ this.state.newFundingScheduleDialogOpen }
        />
        <Button
          className="w-full"
          variant="outlined"
          color="primary"
          onClick={ this.openNewFundingScheduleDialog }
        >
          New Funding Schedule
        </Button>
        <Divider/>
        <List>
          {
            fundingSchedules.map(schedule => (
              <ListItem key={ schedule.fundingScheduleId } button
                        onClick={ this.selectItem(schedule.fundingScheduleId) }>
                <ListItemIcon>
                  <Checkbox
                    edge="start"
                    checked={ selectedFundingSchedule === schedule.fundingScheduleId }
                    tabIndex={ -1 }
                    color="primary"
                    disabled={ !!disabled }
                  />
                </ListItemIcon>
                <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 w-full">
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
)(FundingScheduleSelectionList);

