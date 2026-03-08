import type React from 'react';
import { Fragment, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

import Typography from '@monetr/interface/components/Typography';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import MSidebarToggle from './MSidebarToggle';
import MSpan from './MSpan';
import styles from './MTopNavigation.module.scss';

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

  const titleClass = mergeTailwind(
    styles.title,
    props.breadcrumb && styles.titleBreadcrumb,
    props.base && styles.titleClickable,
  );

  const titleTextClass = mergeTailwind(styles.titleText, props.breadcrumb && styles.titleTextHidden);

  const iconClass = mergeTailwind(styles.icon, props.breadcrumb && styles.iconBreadcrumb);

  return (
    <Fragment>
      <div className={styles.spacer} />
      <div className={styles.topNav}>
        <div className={styles.topNavLeft}>
          <MSidebarToggle backButton={props.base} className='mr-2' />
          <span className={styles.titleWrapper}>
            <Typography className={titleClass} ellipsis onClick={onInitialClick} size='2xl' weight='bold'>
              <Icon className={iconClass} />
              <span className={titleTextClass}>{props.title}</span>
            </Typography>
            {Boolean(props.breadcrumb) && (
              <Fragment>
                <MSpan className={styles.breadcrumbSeparator} color='subtle' size='2xl' weight='bold'>
                  /
                </MSpan>
                <MSpan className={styles.breadcrumbText} color='emphasis' ellipsis size='2xl' weight='bold'>
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

  return <div className={styles.actionArea}>{props.children}</div>;
}
