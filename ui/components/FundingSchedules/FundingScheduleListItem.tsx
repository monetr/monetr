import React from 'react';
import { ChevronRight } from '@mui/icons-material';
import { ListItem, Typography } from '@mui/material';

import { useFundingSchedule } from 'hooks/fundingSchedules';
import { useSpendingSink } from 'hooks/spending';
import formatAmount from 'util/formatAmount';
import getFundingScheduleContribution from 'util/getFundingScheduleContribution';

interface Props {
  fundingScheduleId: number;
}

export default function FundingScheduleListItem(props: Props): JSX.Element {
  const schedule = useFundingSchedule(props.fundingScheduleId);
  const { result: spending } = useSpendingSink();
  const contribution = getFundingScheduleContribution(props.fundingScheduleId, spending);

  return (
    <ListItem key={ schedule.fundingScheduleId } button>
      <div className="grid grid-cols-4 grid-rows-3 grid-flow-col gap-1 w-full">
        <div className="col-span-2">
          <Typography>
            { schedule.name }
          </Typography>
        </div>
        <div className="col-span-3 opacity-50">
          <Typography variant="body2">
            { schedule.description }
          </Typography>
        </div>
        <div className="col-span-3 opacity-70">
          <Typography variant="body2">
            Contribution: { formatAmount(contribution) }
          </Typography>
        </div>
        <div className="col-span-1 flex justify-end">
          <Typography variant="subtitle2" color="primary">
            { schedule.nextOccurrence.format('MMM Do') }
          </Typography>
        </div>
        <div className="row-span-3 col-span-1 flex justify-end items-center">
          <ChevronRight />
        </div>
      </div>
    </ListItem>
  );
}
