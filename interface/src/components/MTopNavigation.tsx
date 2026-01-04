import type React from 'react';
import { Fragment, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import MSidebarToggle from './MSidebarToggle';
import MSpan from './MSpan';

export interface MTopNavigationProps {
  icon: React.FC<{ className?: string }>;
  title: string;
  children?: React.ReactNode;
  base?: string;
  breadcrumb?: string;
}

export default function MTopNavigation(props: MTopNavigationProps): JSX.Element {
  const Icon = props.icon;
  const navigate = useNavigate();

  const onInitialClick = useCallback(() => {
    if (props.base) {
      navigate(props.base);
    }
  }, [props.base, navigate]);

  const className = mergeTailwind(
    {
      'dark:text-dark-monetr-content-emphasis': !props.breadcrumb,
      'dark:text-dark-monetr-content-subtle dark:hover:text-dark-monetr-content-emphasis': Boolean(props.breadcrumb),
      'cursor-pointer': Boolean(props.base),
    },
    'w-auto order-1 flex-shrink-0 md:flex-shrink',
  );

  const titleClassName = mergeTailwind(
    {
      'hidden md:inline': Boolean(props.breadcrumb),
    },
    'w-auto text-center order-1',
  );

  const iconClassName = mergeTailwind('mb-1 inline', {
    'mr-0 md:mr-2': Boolean(props.breadcrumb),
    'mr-2': !props.breadcrumb,
  });

  return (
    <Fragment>
      <div className='pb-12' />
      <div className='h-auto w-full md:h-12 flex flex-col md:flex-row md:items-center px-4 gap-x-2 justify-between fixed z-30 top-0 backdrop-blur-sm bg-gradient-to-t from-transparent dark:to-dark-monetr-background via-90%'>
        <div className='flex gap-2 min-w-0 h-12 items-center flex-shrink'>
          <MSidebarToggle backButton={props.base} className='mr-2' />
          <span className='flex gap-2 flex-grow min-w-0'>
            <MSpan className={className} ellipsis onClick={onInitialClick} size='2xl' weight='bold'>
              <Icon className={iconClassName} />
              <span className={titleClassName}>{props.title}</span>
            </MSpan>
            {Boolean(props.breadcrumb) && (
              <Fragment>
                <MSpan className='hidden md:block order-2' color='subtle' size='2xl' weight='bold'>
                  /
                </MSpan>
                <MSpan className='order-3' color='emphasis' ellipsis size='2xl' weight='bold'>
                  {props.breadcrumb}
                </MSpan>
              </Fragment>
            )}
          </span>
        </div>
      </div>
      <ActionArea>{props.children}</ActionArea>
    </Fragment>
  );
}

interface ActionAreaProps {
  children?: React.ReactNode;
}

function ActionArea(props: ActionAreaProps): JSX.Element {
  if (!props.children) {
    return null;
  }

  const styles = mergeTailwind(
    'flex justify-end gap-x-4 md:gap-x-2 z-40',
    'flex-shrink-0',
    'fixed -bottom-1 top-[unset] md:top-2 md:bottom-[unset] left-0 md:left-[unset] md:right-4',
    // Hacky width and padding to make sure scrollbar renders properly
    'p-6 pt-2 pr-[calc(1.5rem-16px)] md:p-0 pb-8 md:pb-0',
    'w-[calc(100vw-16px)] md:w-auto',
    'backdrop-blur-sm bg-dark-monetr-background/50',
  );

  return <div className={styles}>{props.children}</div>;
}
