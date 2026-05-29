import type { HTMLAttributes } from 'react';

import Flex, { type FlexProps } from '@monetr/interface/components/Flex';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Item.module.scss';

export type ItemProps = HTMLAttributes<HTMLLIElement>;

export function Item({ className, ...props }: ItemProps): React.JSX.Element {
  return (
    <li className={mergeClasses(styles.itemRoot, className)} {...props}>
      {props.children}
    </li>
  );
}

export type ItemContentProps = FlexProps;

export function ItemContent(props: ItemContentProps): React.JSX.Element {
  return <Flex align='center' flex='grow' justify='end' shrink='none' width='fit' {...props} />;
}
