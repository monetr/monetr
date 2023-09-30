import React, { Fragment, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

import MSidebarToggle from './MSidebarToggle';
import MSpan from './MSpan';
import { ReactElement } from './types';

import mergeTailwind from 'util/mergeTailwind';

export interface MTopNavigationProps {
  icon: React.FC<{ className?: string }>;
  title: string;
  children?: ReactElement;
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

  const className = mergeTailwind({
    'dark:text-dark-monetr-content-emphasis': !Boolean(props.breadcrumb),
    'dark:text-dark-monetr-content-subtle dark:hover:text-dark-monetr-content-emphasis': Boolean(props.breadcrumb),
    'cursor-pointer': Boolean(props.base),
  }, 'w-auto order-1');

  const titleClassName = mergeTailwind({
    'hidden md:inline': Boolean(props.breadcrumb),
  }, 'w-auto text-center order-1');

  const iconClassName = mergeTailwind({
    'mr-0 md:mr-2': Boolean(props.breadcrumb),
    'mr-2': !Boolean(props.breadcrumb),
  }, 'mb-1');

  function InitialCrumb(): JSX.Element {
    return (
      <MSpan weight='bold' size='2xl' className={ className } onClick={ onInitialClick } ellipsis>
        <Icon className={ iconClassName } />
        <span className={ titleClassName }>
          { props.title }
        </span>
      </MSpan>
    );
  }

  function BreadcrumbMaybe(): JSX.Element {
    if (!props.breadcrumb) return null;

    return (
      <Fragment>
        <MSpan weight='bold' size='2xl' color='subtle' className='hidden md:block order-2'>
          /
        </MSpan>
        <MSpan weight='bold' size='2xl' color='emphasis' ellipsis className='order-3'>
          { props.breadcrumb }
        </MSpan>
      </Fragment>
    );
  }

  return (
    <div className='w-full h-auto md:h-12 flex flex-col md:flex-row md:items-center px-4 gap-4 justify-between'>
      <div className='flex gap-2 min-w-0 h-12 items-center flex-grow'>
        <MSidebarToggle className='mr-2' backButton={ props.base } />
        <span className='flex gap-2 flex-grow min-w-0'>
          <InitialCrumb />
          <BreadcrumbMaybe />
        </span>
      </div>
      { props.children && (
        <div className='flex gap-2 pb-4 md:p-0'>
          { props.children }
        </div>
      )}
    </div>
  );
}
