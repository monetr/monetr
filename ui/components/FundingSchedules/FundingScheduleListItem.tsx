import React, { Fragment, useMemo } from 'react';
import { MoreVert, AttachMoney } from '@mui/icons-material';
import { Divider, IconButton, ListItem } from '@mui/material';
import moment from 'moment';

import { useFundingSchedule } from 'hooks/fundingSchedules';
import { useSpendingSink } from 'hooks/spending';
import formatAmount from 'util/formatAmount';
import getColor from 'util/getColor';
import getFundingScheduleContribution from 'util/getFundingScheduleContribution';

interface Props {
  fundingScheduleId: number;
}

export default function FundingScheduleListItem(props: Props): JSX.Element {
  const schedule = useFundingSchedule(props.fundingScheduleId);
  const { result: spending } = useSpendingSink();
  const contribution = getFundingScheduleContribution(props.fundingScheduleId, spending);
  const color = useMemo(() => getColor(schedule.name), [schedule.name]);

  const next = schedule.nextOccurrence;
  const dateFormatString = next.year() !== moment().year() ? 'dddd MMMM Do, yyyy' : 'dddd MMMM Do';
  const nextOccurrenceString = `${ next.format(dateFormatString) } (${ next.fromNow() })`;

  return (
    <Fragment>
      <ListItem>
        <div className="flex flex-row gap-2 h-16 w-full mt-1 mb-1">
          <div className="rounded-lg flex w-16" style={ { backgroundColor: `${color}` } }>
            <AttachMoney className="col-span-1 h-16 w-10 m-auto fill-gray-500" />
          </div>
          <div className="flex h-full flex-col">
            <span className="text-2xl font-semibold mt-auto text-gray-700">
              { schedule.name }
            </span>
            <span className="text-xl font-normal mt-auto text-gray-400">
              { nextOccurrenceString }
            </span>
          </div>
          <div className="flex-grow flex h-full flex-col items-end justify-center">
            <span className="text-2xl font-normal text-gray-500">
              { formatAmount(contribution) }
            </span>
          </div>
          <div className="flex h-full flex-col items-center justify-center">
            <IconButton>
              <MoreVert />
            </IconButton>
          </div>
        </div>
      </ListItem>
      <Divider />
    </Fragment>
  );
}
