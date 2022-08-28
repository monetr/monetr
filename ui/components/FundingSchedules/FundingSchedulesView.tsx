import { Divider, List } from '@mui/material';
import FundingScheduleListItem from 'components/FundingSchedules/FundingScheduleListItem';
import { useFundingSchedules } from 'hooks/fundingSchedules';
import React, { Fragment } from 'react';


export default function FundingSchedulesView(): JSX.Element {
  const fundingSchedules = useFundingSchedules();
  return (

    <div className="minus-nav">
      <div className="w-full view-area bg-white">
        <List className="w-full pt-0" dense>
          {
            Array.from(fundingSchedules.values())
              .map(schedule => (
                <Fragment
                  key={ schedule.fundingScheduleId }
                >
                  <FundingScheduleListItem
                    fundingScheduleId={ schedule.fundingScheduleId }
                  />
                  <Divider />
                </Fragment>
              ))
          }
        </List>
      </div>
    </div>
  );
}
