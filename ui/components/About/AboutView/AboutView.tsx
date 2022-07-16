import React, { Fragment } from 'react';
import { GitHub } from '@mui/icons-material';
import { Button } from '@mui/material';

import CodeBasic from 'components/Code/CodeBasic/CodeBasic';
import { useAppConfiguration } from 'hooks/useAppConfiguration';

export default function AboutView(): JSX.Element {
  const {
    release,
    revision,
    buildType,
    buildTime,
  } = useAppConfiguration();

  function Version(): JSX.Element {
    if (!release) {
      return null;
    }

    return (
      <Fragment>
        <div className="grid grid-cols-2">
          <div className="flex items-center">
            <span className="text-lg">Version</span>
          </div>
          <div className="flex items-center">
            <CodeBasic>{ release }</CodeBasic>
            <Button
              target="_blank"
              href={ `https://github.com/monetr/monetr/releases/tag/${ release }` }
              variant="outlined"
              color="inherit"
              className="ml-auto"
            >
              <GitHub className="mr-2.5" />
              Release Notes
            </Button>
          </div>
        </div>
        <hr />
      </Fragment>
    );
  }

  function Revision(): JSX.Element {
    return (
      <Fragment>
        <div className="grid grid-cols-2">
          <div className="flex items-center">
            <span className="text-lg">Git Revision</span>
          </div>
          <div className="flex items-center">
            <CodeBasic>{ revision.slice(0, 7) }</CodeBasic>
            <Button
              target="_blank"
              href={ `https://github.com/monetr/monetr/tree/${ revision }` }
              variant="outlined"
              color="inherit"
              className="ml-auto"
            >
              <GitHub className="mr-2.5" />
              Browse
            </Button>
          </div>
        </div>
        <hr />
      </Fragment>
    );
  }

  function BuildType(): JSX.Element {
    return (
      <Fragment>
        <div className="grid grid-cols-2">
          <div className="flex items-center">
            <span className="text-lg">Build Type</span>
          </div>
          <div className="flex items-center">
            <CodeBasic>{ buildType }</CodeBasic>
          </div>
        </div>
        <hr />
      </Fragment>
    );
  }

  function BuildTime(): JSX.Element {
    if (!buildTime) {
      return null;
    }

    return (
      <Fragment>
        <div className="grid grid-cols-2">
          <div className="flex items-center">
            <span className="text-lg">Build Time</span>
          </div>
          <div className="flex items-center">
            <CodeBasic>
              { buildTime.format('MMMM Do YYYY, h:mma Z') }
            </CodeBasic>
          </div>
        </div>
        <hr />
      </Fragment>
    );
  }

  return (
    <div className="grid gap-5 w-full lg:w-2/3">
      <div>
        <span className="text-2xl mb-2.5">
          About monetr
        </span>
        <div className="grid mt-2.5 gap-2.5">
          <Version />
          <Revision />
          <BuildType />
          <BuildTime />
        </div>
      </div>
      <div>
        <span className="text-2xl mb-2.5">
          Getting Help
        </span>
        <div className="grid mt-2.5 gap-2.5">
          <div className="grid grid-cols-2">
            <div className="flex items-center">
              <span className="text-lg">Email</span>
            </div>
            <div className="flex items-center">
              <a
                target="_blank"
                href="mailto:support@monetr.app"
              >
                support@monetr.app
              </a>
            </div>
          </div>
          <hr />
          <div className="grid grid-cols-2">
            <div className="flex items-center">
              <span className="text-lg">GitHub Discussions</span>
            </div>
            <div className="flex items-center">
              <a
                target="_blank"
                href="https://github.com/monetr/monetr/discussions"
              >
                https://github.com/monetr/monetr/discussions
              </a>
            </div>
          </div>
          <hr />
        </div>
      </div>
    </div>
  );
}
