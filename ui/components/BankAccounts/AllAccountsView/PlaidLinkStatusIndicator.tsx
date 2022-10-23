import { FiberManualRecord } from "@mui/icons-material";
import { styled, Tooltip, tooltipClasses, TooltipProps } from "@mui/material";
import { useInstitution } from "hooks/institutions";
import Link, { LinkStatus } from "models/Link";
import React, { Fragment } from "react"

interface Props {
  link: Link;
}

export default function PlaidLinkStatusIndicator(props: Props): JSX.Element {
  if (!props.link.getIsPlaid()) {
    throw new Error('Cannot show a Plaid link status for a non-Plaid link.')
  }

  const { result: institution } = useInstitution(props.link.plaidInstitutionId);

  // Build the initial color for the status indicator based on the status of the link.
  let colorClassName = {
    [LinkStatus.Setup]: 'text-green-500',
    [LinkStatus.Pending]: 'text-yellow-500',
    [LinkStatus.Error]: 'text-red-500',
    [LinkStatus.Unknown]: 'text-gray-500',
  }[props.link.linkStatus];

  // But once the institution has data loaded, check if we are getting transaction updates from Plaid.
  if (institution?.status) {
    if (props.link.linkStatus === LinkStatus.Setup && institution.status.transactions_updates?.status !== 'HEALTHY') {
      // If we are not, then change the status indicator to yellow.
      colorClassName = 'text-yellow-500';
    }
  }

  const plaidStatusString = {
    ['HEALTHY']: 'Healthy',
    ['DEGRADED']: 'Having problems',
    ['DOWN']: 'Offline',
  }

  const statusString = institution?.status?.transactions_updates?.status ?
    plaidStatusString[institution.status.transactions_updates.status] :
    'Unknown';

  const NoMaxWidthTooltip = styled(({ className, ...props }: TooltipProps ) => (
    <Tooltip {...props} classes={{ popper: className }} />
  ))({
    [`& .${tooltipClasses.tooltip}`]: {
      maxWidth: 'none',
    },
  });

  function TooltipInner(): JSX.Element {
    return (
      <Fragment>
        <div className='m-2 text-base'>
          <ul><b>Last Successful Sync:</b> <span>{ props.link.lastSuccessfulUpdate ? props.link.lastSuccessfulUpdate.format('MMMM Do, h:mm a') : 'N/A' }</span></ul>
          <ul><b>Transaction & Balance Updates:</b> <span>{ statusString }</span></ul>
        </div>
      </Fragment>
    );
  }

  return (
    <NoMaxWidthTooltip title={ <TooltipInner /> }>
      <FiberManualRecord className={ `mr-2 ${colorClassName}` } />
    </NoMaxWidthTooltip>
  );
}
