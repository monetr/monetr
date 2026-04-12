import type { ReactNode } from 'react';

import styles from './Cards.module.scss';

import { Link } from '@rspress/core/theme-original';

interface CardProps {
  title: string;
  href: string;
  icon?: ReactNode;
  children?: ReactNode;
  arrow?: boolean;
}

function Card({ title, href, icon, children, arrow }: CardProps) {
  return (
    <Link className={styles.card} href={href}>
      {icon && <span className={styles.icon}>{icon}</span>}
      <h3 className={styles.title}>
        {title}
        {arrow && <span className={styles.titleArrow}>&rarr;</span>}
      </h3>
      {children && <p className={styles.description}>{children}</p>}
    </Link>
  );
}

interface CardsProps {
  children: ReactNode;
  num?: number;
}

export function Cards({ children, num }: CardsProps) {
  return (
    <div
      className={styles.cardsGrid}
      style={{
        gridTemplateColumns: `repeat(${num ?? 2}, minmax(0, 1fr))`,
      }}
    >
      {children}
    </div>
  );
}

Cards.Card = Card;
