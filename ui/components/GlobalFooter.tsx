import { Typography } from '@mui/material';
import React, { Fragment } from 'react';
import { useSelector } from 'react-redux';
import { getRelease } from 'shared/bootstrap/selectors';

const GlobalFooter = (): JSX.Element => {
  const release = useSelector(getRelease);

  let versionLink: JSX.Element = null;
  if (release) {
    versionLink = (
      <Fragment>
        <span>- </span>
        <a
          target="_blank"
          href={ `https://github.com/monetr/monetr/releases/tag/${ release }` }
        >
          { release }
        </a>
      </Fragment>
    )
  }

  return (
    <Typography
      className="absolute inline w-full text-center bottom-1 opacity-30"
    >
      © { new Date().getFullYear() } monetr LLC { versionLink }
    </Typography>
  );
};

export default GlobalFooter;