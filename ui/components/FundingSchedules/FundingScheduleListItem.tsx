import React, { Fragment, useMemo, useState } from 'react';
import { AttachMoney, MoreVert, Remove, Weekend } from '@mui/icons-material';
import { Divider, IconButton, ListItem, Menu, MenuItem, Skeleton } from '@mui/material';
import moment from 'moment';

import { showRemoveFundingScheduleDialog } from './RemoveFundingScheduleDialog';

import clsx from 'clsx';
import { useCurrentBalance } from 'hooks/balances';
import { useNextFundingForecast } from 'hooks/forecast';
import { useFundingSchedule, useUpdateFundingSchedule } from 'hooks/fundingSchedules';
import FundingSchedule from 'models/FundingSchedule';
import formatAmount from 'util/formatAmount';
import getColor from 'util/getColor';

interface Props {
  fundingScheduleId: number;
}

export default function FundingScheduleListItem(props: Props): JSX.Element {
  const [menuAnchor, setMenuAnchor] = useState<Element | null>();
  const openMenu = (event: { currentTarget: Element }) => setMenuAnchor(event.currentTarget);
  const closeMenu = () => setMenuAnchor(null);

  const schedule = useFundingSchedule(props.fundingScheduleId);
  const balance = useCurrentBalance();
  const updateFundingSchedule = useUpdateFundingSchedule();
  const contributionForecast = useNextFundingForecast(props.fundingScheduleId);

  const color = useMemo(() => getColor(schedule.name), [schedule.name]);

  const next = schedule.nextOccurrence;
  const dateFormatString = next.year() !== moment().year() ? 'dddd MMMM Do, yyyy' : 'dddd MMMM Do';
  const nextOccurrenceString = `${next.format(dateFormatString)} (${next.fromNow()})`;

  async function updateWeekends(excludeWeekends: boolean) {
    const updatedFunding = new FundingSchedule({
      ...schedule,
      excludeWeekends,
    });

    return updateFundingSchedule(updatedFunding);
  }

  if (!schedule) {
    return null;
  }

  function removeDialog() {
    closeMenu();
    showRemoveFundingScheduleDialog({
      fundingSchedule: schedule,
    });
  }

  function Contribution(): JSX.Element {
    if (contributionForecast.isLoading) {
      return (
        // TODO This will break with the next MUI upgrade.
        <Skeleton variant="text" width={ 80 } height={ 24 } />
      );
    }

    if (contributionForecast.result) {
      return (
        <Fragment>
          {formatAmount(contributionForecast.result)}
        </Fragment>
      );
    }

    return (
      <Fragment>
        N/A
      </Fragment>
    );
  }

  function EstimatedSafeToSpend(): JSX.Element {
    if (!schedule.estimatedDeposit) {
      return null;
    }

    const loader = <Skeleton variant="text" width={ 80 } height={ 24 } />;

    let textColor = 'text-gray-500';

    let amount: string | null;
    if (!contributionForecast.isLoading && balance !== null) {
      const est = (balance.free + schedule.estimatedDeposit) - contributionForecast.result;
      amount = formatAmount(est);
      textColor = est > 0 ? 'text-green-500' : 'text-red-500';
    }

    return (
      <div className="flex-grow flex h-full flex-col items-end justify-center">
        <span className="font-normal text-gray-500 text-lg">
          Estimated Free-To-Use
        </span>
        <span className={ clsx('text-md font-normal', textColor) }>
          {amount || loader}
        </span>
      </div>
    );
  }

  return (
    <Fragment>
      <ListItem>
        <div className="flex flex-row gap-2 h-14 w-full mt-1 mb-1">
          <div className="rounded-lg flex w-14" style={ { backgroundColor: `${color}` } }>
            <AttachMoney className="col-span-1 h-14 w-10 m-auto fill-gray-500" />
          </div>
          <div className="flex h-full flex-col">
            <span className="font-semibold mt-auto text-gray-700 text-xl">
              {schedule.name}
            </span>
            <span className="font-normal mt-auto text-gray-400 text-md">
              {nextOccurrenceString}
            </span>
          </div>
          <EstimatedSafeToSpend />
          <div className="flex-grow flex h-full flex-col items-end justify-center">
            <span className="font-normal text-gray-500 text-lg">
              Next Contribution
            </span>
            <span className="font-normal text-gray-500 text-md">
              <Contribution />
            </span>
          </div>
          <div className="flex h-full flex-col items-center justify-center">
            <IconButton onClick={ openMenu }>
              <MoreVert />
            </IconButton>
            <Menu
              id={ `${schedule.fundingScheduleId}-menu` }
              anchorEl={ menuAnchor }
              keepMounted
              open={ Boolean(menuAnchor) }
              onClose={ closeMenu }
            >
              <MenuItem onClick={ () => updateWeekends(!schedule.excludeWeekends) }>
                <Weekend className="mr-2" />
                {schedule.excludeWeekends ? 'Include weekends' : 'Exclude weekends'}
              </MenuItem>
              <MenuItem
                className="text-red-500"
                onClick={ removeDialog }
              >
                <Remove className="mr-2" />
                Remove
              </MenuItem>
            </Menu>
          </div>
        </div>
      </ListItem>
      <Divider />
    </Fragment>
  );
}
