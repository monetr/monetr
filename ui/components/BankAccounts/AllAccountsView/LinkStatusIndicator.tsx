import React from 'react';
import { FiberManualRecord } from '@mui/icons-material';
import { Tooltip } from '@mui/material';

import Link, { LinkStatus } from 'models/Link';

interface Props {
  link: Link;
}

export default function LinkStatusIndicator(props: Props): JSX.Element {
  switch (props.link.linkStatus) {
    case LinkStatus.Setup:
      return (
        <Tooltip title="This link is working properly.">
          <FiberManualRecord className="mr-2 text-green-500" />
        </Tooltip>
      );
    case LinkStatus.Pending:
      return (
        <Tooltip title="This link has not been completely setup yet.">
          <FiberManualRecord className="mr-2 text-yellow-500" />
        </Tooltip>
      );
    case LinkStatus.Error:
      return (
        <Tooltip title={ props.link.getErrorMessage() }>
          <FiberManualRecord className="mr-2 text-red-500" />
        </Tooltip>
      );
    case LinkStatus.Unknown:
      return <FiberManualRecord className="mr-2 text-gray-500" />;
  }
}
