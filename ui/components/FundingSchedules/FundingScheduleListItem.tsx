import React, { Fragment, useMemo, useState } from 'react';
import { MoreVert, AttachMoney, Remove, Weekend } from '@mui/icons-material';
import { Divider, IconButton, ListItem, Menu, MenuItem, Skeleton } from '@mui/material';
import moment from 'moment';

import { useFundingSchedule, useUpdateFundingSchedule } from 'hooks/fundingSchedules';
import formatAmount from 'util/formatAmount';
import getColor from 'util/getColor';
import FundingSchedule from 'models/FundingSchedule';
import { showRemoveFundingScheduleDialog } from './RemoveFundingScheduleDialog';
import { useNextFundingForecast } from 'hooks/forecast';

interface Props {
  fundingScheduleId: number;
}

export default function FundingScheduleListItem(props: Props): JSX.Element {
  const [menuAnchor, setMenuAnchor] = useState<Element | null>();
  const openMenu = (event: { currentTarget: Element }) => setMenuAnchor(event.currentTarget);
  const closeMenu = () => setMenuAnchor(null);

  const schedule = useFundingSchedule(props.fundingScheduleId);
  const updateFundingSchedule = useUpdateFundingSchedule();
  const contributionForecast = useNextFundingForecast(props.fundingScheduleId);

  const color = useMemo(() => getColor(schedule.name), [schedule.name]);

  const next = schedule.nextOccurrence;
  const dateFormatString = next.year() !== moment().year() ? 'dddd MMMM Do, yyyy' : 'dddd MMMM Do';
  const nextOccurrenceString = `${ next.format(dateFormatString) } (${ next.fromNow() })`;

  async function updateWeekends(excludeWeekends: boolean){
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
        <Skeleton variant="text" width={80} height={28} />
      );
    }

    if (contributionForecast.result) {
      return (
        <Fragment>
          { formatAmount(contributionForecast.result) }
        </Fragment>
      )
    }

    return (
      <Fragment>
        N/A
      </Fragment>
    )
  }

  return (
    <Fragment>
      <ListItem>
        <div className="flex flex-row gap-2 h-16 w-full mt-1 mb-1">
          <div className="rounded-lg flex w-16" style={ { backgroundColor: `${color}` } }>
            <AttachMoney className="col-span-1 h-16 w-10 m-auto fill-gray-500" />
          </div>
          <div className="flex h-full flex-col">
            <span className="sm:text-2xl font-semibold mt-auto text-gray-700 text-lg">
              { schedule.name }
            </span>
            <span className="sm:text-xl font-normal mt-auto text-gray-400 text-md">
              { nextOccurrenceString }
            </span>
          </div>
          <div className="flex-grow flex h-full flex-col items-end justify-center">
            <span className="sm:text-xl font-normal text-gray-500 text-md">
              Next Contribution
            </span>
            <span className="sm:text-lg font-normal text-gray-500 text-sm">
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
                { schedule.excludeWeekends ? 'Include weekends' : 'Exclude weekends' }
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
