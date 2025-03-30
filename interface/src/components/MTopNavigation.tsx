import React, { Fragment, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

import MSidebarToggle from './MSidebarToggle';
import MSpan from './MSpan';
import { ReactElement } from './types';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

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
  }, 'w-auto order-1 flex-shrink-0 md:flex-shrink');

  const titleClassName = mergeTailwind({
    'hidden md:inline': Boolean(props.breadcrumb),
  }, 'w-auto text-center order-1');

  const iconClassName = mergeTailwind('mb-1 inline', {
    'mr-0 md:mr-2': Boolean(props.breadcrumb),
    'mr-2': !Boolean(props.breadcrumb),
  });

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
    <div className='w-full h-auto md:h-12 flex flex-col md:flex-row md:items-center px-4 gap-x-2 justify-between'>
      <div className='flex gap-2 min-w-0 h-12 items-center flex-shrink'>
        <MSidebarToggle className='mr-2' backButton={ props.base } />
        <span className='flex gap-2 flex-grow min-w-0'>
          <InitialCrumb />
          <BreadcrumbMaybe />
        </span>
      </div>
      <ActionArea children={ props.children } />
    </div>
  );
}

interface ActionAreaProps {
  children?: React.ReactNode;
}

function ActionArea(props: ActionAreaProps): JSX.Element {
  if (!props.children) return null;

  const styles = mergeTailwind(
    'flex justify-end gap-x-4 md:gap-x-2',
    'flex-shrink-0',
    'md:relative fixed -bottom-1 md:bottom-auto left-0 md:left-auto',
    // Hacky width and padding to make sure scrollbar renders properly
    'p-6 pr-[calc(1.5rem-16px)] md:p-0 pb-8 md:pb-0',
    'w-[calc(100vw-16px)] md:w-auto',
    'z-20',
    'backdrop-blur-sm bg-dark-monetr-background/50',
  );

  return (
    <div className={ styles }>
      { props.children }
    </div>
  );
}
