import React, { useEffect, useState } from 'react';
import { Button, Checkbox, Divider, List, ListItem, ListItemIcon, Typography } from '@mui/material';

import { showCreateFundingScheduleDialog } from 'components/FundingSchedules/CreateFundingScheduleDialog';
import { useFundingSchedules } from 'hooks/fundingSchedules';
import FundingSchedule from 'models/FundingSchedule';

interface Props {
  initialValue?: number;
  onChange: (_fundingSchedule: FundingSchedule) => void;
  disabled?: boolean;
}

export default function FundingScheduleSelectionList(props: Props): JSX.Element {
  const [selectedFundingSchedule, setSelectedFundingSchedule] = useState<FundingSchedule | null>(null);

  const fundingSchedules = useFundingSchedules();

  useEffect(() => {
    if (props.initialValue) setSelectedFundingSchedule(fundingSchedules.get(props.initialValue));
  }, [setSelectedFundingSchedule, props.initialValue, fundingSchedules]);

  function selectItem(fundingScheduleId: number) {
    const fundingSchedule = fundingSchedules.get(fundingScheduleId);
    setSelectedFundingSchedule(fundingSchedule);
    props.onChange(fundingSchedule);
  }

  return (
    <div className="w-full funding-schedule-selection-list">
      <Button
        className="w-full mb-2.5"
        variant="outlined"
        color="primary"
        onClick={ showCreateFundingScheduleDialog }
      >
        New Funding Schedule
      </Button>
      <Divider />
      <List>
        {
          Array.from(fundingSchedules.values())
            .map(schedule => (
              <ListItem key={ schedule.fundingScheduleId } button
                onClick={ () => selectItem(schedule.fundingScheduleId) }>
                <ListItemIcon>
                  <Checkbox
                    edge="start"
                    checked={ selectedFundingSchedule?.fundingScheduleId === schedule.fundingScheduleId }
                    tabIndex={ -1 }
                    color="primary"
                    disabled={ !!props.disabled }
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
            ))
        }
      </List>
    </div>
  );
}
