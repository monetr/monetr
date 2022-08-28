import React from 'react';
import { ChevronRight } from '@mui/icons-material';
import { ListItem, Typography } from '@mui/material';

import { useFundingSchedule } from 'hooks/fundingSchedules';
import { useSpendingSink } from 'hooks/spending';
import formatAmount from 'util/formatAmount';
import getFundingScheduleContribution from 'util/getFundingScheduleContribution';
import AttachMoneyIcon from '@mui/icons-material/AttachMoney';
import getColor from 'util/getColor';
import classNames from 'classnames';
import moment from 'moment';

interface Props {
  fundingScheduleId: number;
}

export default function FundingScheduleListItem(props: Props): JSX.Element {
  const schedule = useFundingSchedule(props.fundingScheduleId);
  const { result: spending } = useSpendingSink();
  const contribution = getFundingScheduleContribution(props.fundingScheduleId, spending);
  const color = getColor(schedule.name);

  const next = schedule.nextOccurrence;
  const dateFormatString = next.year() !== moment().year() ? 'dddd MMMM Do, yyyy' : 'dddd MMMM Do';
  const nextOccurrenceString = `${ next.format(dateFormatString) } (${ next.fromNow() })`;

  return (
    <ListItem key={ schedule.fundingScheduleId } button>
      <div className="grid grid-cols-12 h-16 w-full mt-1 mb-1">
        <div className="col-span-1 rounded-lg flex w-16 bg-gray-200">
          <AttachMoneyIcon className="col-span-1 h-16 w-10 m-auto fill-gray-500" />
        </div>
        <div className="lg:col-span-4 flex h-full flex-col">
          <span className="text-3xl font-semibold mt-auto text-gray-700">
            { schedule.name }
          </span>
          <span className="text-xl font-normal mt-auto text-gray-400">
            { nextOccurrenceString }
          </span>
        </div>
        <div className="col-span-2">

        </div>
      </div>
    </ListItem>
  );
}
