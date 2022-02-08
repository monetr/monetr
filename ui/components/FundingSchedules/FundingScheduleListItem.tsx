import { ChevronRight } from '@mui/icons-material';
import { IconButton, ListItem, Typography } from '@mui/material';
import React from 'react';
import { useSelector } from 'react-redux';
import { getFundingScheduleById } from 'shared/fundingSchedules/selectors/getFundingScheduleById';
import { getFundingScheduleContribution } from 'shared/fundingSchedules/selectors/getFundingScheduleContribution';
import formatAmount from 'util/formatAmount';

interface PropTypes {
  fundingScheduleId: number;
}

export default function FundingScheduleListItem(props: PropTypes): JSX.Element {
  const schedule = useSelector(getFundingScheduleById(props.fundingScheduleId));
  const contribution = useSelector(getFundingScheduleContribution(props.fundingScheduleId));

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
        <div className="row-span-3 col-span-1 flex justify-end">
          <IconButton>
            <ChevronRight/>
          </IconButton>
        </div>
      </div>
    </ListItem>
  );
}
