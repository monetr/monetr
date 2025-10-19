import type React from 'react';
import { twMerge } from 'tailwind-merge';

import Link from 'next/link';

interface FeatureProps {
  title: React.ReactNode;
  description?: React.ReactNode;
  className?: string;
  link?: string;
  linkText?: React.ReactNode;
  linkExternal?: boolean;
}

export default function Feature(props: FeatureProps): JSX.Element {
  const className = twMerge(
    'rounded-3xl bg-black border border-zinc-700 bg-opacity-20 backdrop-blur-sm shadow-lg flex flex-col overflow-hidden  justify-between',
    props.className,
  );

  return (
    <div className={className}>
      <div className='w-full p-8 flex flex-col justify-evenly gap-4'>
        {props.title}
        {props.description && props.description}
      </div>
      {props.link && (
        <Link
          href={props.link}
          target={props.linkExternal ? '_blank' : undefined}
          className='w-full bottom-0 block px-8 py-3 text-md font-semibold text-center text-gray-100 transition duration-100 bg-white outline-none bg-opacity-10 hover:bg-opacity-20 md:text-base'
        >
          {props.linkText ?? 'Learn More'}
        </Link>
      )}
    </div>
  );
}
