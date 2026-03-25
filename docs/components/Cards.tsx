import type { ReactNode } from 'react';

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
    <Link
      className='block p-4 rounded-lg border border-zinc-700 hover:border-zinc-500 transition-colors bg-black bg-opacity-20 backdrop-blur-sm no-underline text-inherit'
      href={href}
    >
      {icon && <span className='text-2xl mb-2 block'>{icon}</span>}
      <h3 className='text-lg font-semibold'>
        {title}
        {arrow && <span className='ml-1'>&rarr;</span>}
      </h3>
      {children && <p className='text-sm text-zinc-400 mt-1'>{children}</p>}
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
      className='grid gap-4 mt-4'
      style={{
        gridTemplateColumns: `repeat(${num ?? 2}, minmax(0, 1fr))`,
      }}
    >
      {children}
    </div>
  );
}

Cards.Card = Card;
