import React from 'react';

import { twMerge } from 'tailwind-merge';

export interface CardProps extends React.HTMLProps<HTMLDivElement> {}

export default function Card({ className, ...props }: CardProps): JSX.Element {
  return <div className={twMerge('border-dark-monetr-border rounded border p-4 space-y-2', className)} {...props} />;
}
