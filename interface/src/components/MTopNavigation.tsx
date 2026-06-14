import type React from 'react';
import { Fragment, useCallback } from 'react';
import { useLocation } from 'wouter';

import MSidebarToggle from '@monetr/interface/components/MSidebarToggle';
import Typography from '@monetr/interface/components/Typography';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './MTopNavigation.module.scss';

export interface MTopNavigationProps {
  icon: React.FC<{ className?: string }>;
  title: string;
  children?: React.ReactNode;
  base?: string;
  breadcrumb?: string;
}

export default function MTopNavigation(props: MTopNavigationProps): React.JSX.Element {
  const Icon = props.icon;
  const [, navigate] = useLocation();

  const onInitialClick = useCallback(() => {
    if (props.base) {
      navigate(props.base);
    }
  }, [props.base, navigate]);

  const titleClass = mergeClasses(
    styles.title,
    props.breadcrumb && styles.titleBreadcrumb,
    props.base && styles.titleClickable,
  );

  const titleTextClass = mergeClasses(styles.titleText, props.breadcrumb && styles.titleTextHidden);

  const iconClass = mergeClasses(styles.icon, props.breadcrumb && styles.iconBreadcrumb);

  return (
    <Fragment>
      <div className={styles.spacer} />
      <div className={styles.topNav}>
        <div className={styles.topNavLeft}>
          <MSidebarToggle backButton={props.base} className={styles.toggle} />
          <span className={styles.titleWrapper}>
            <Typography className={titleClass} ellipsis onClick={onInitialClick} size='2xl' weight='bold'>
              <Icon className={iconClass} />
              <span className={titleTextClass}>{props.title}</span>
            </Typography>
            {Boolean(props.breadcrumb) && (
              <Fragment>
                <Typography className={styles.breadcrumbSeparator} color='subtle' size='2xl' weight='bold'>
                  /
                </Typography>
                <Typography className={styles.breadcrumbText} color='emphasis' ellipsis size='2xl' weight='bold'>
                  {props.breadcrumb}
                </Typography>
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

function ActionArea(props: ActionAreaProps): React.ReactNode {
  if (!props.children) {
    return null;
  }

  return <div className={styles.actionArea}>{props.children}</div>;
}
