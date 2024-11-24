import React from 'react';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export const BentoGrid = ({
  className,
  children,
}: {
  className?: string;
  children?: React.ReactNode;
}) => {
  return (
    <div
      className={ mergeTailwind(
        'grid md:auto-rows-[18rem] grid-cols-1 md:grid-cols-3 gap-4 max-w-7xl mx-auto ',
        className
      ) }
    >
      {children}
    </div>
  );
};

export const BentoGridItem = ({
  className,
  title,
  description,
  header,
  icon,
}: {
  className?: string;
  title?: string | React.ReactNode;
  description?: string | React.ReactNode;
  header?: React.ReactNode;
  icon?: React.ReactNode;
}) => {
  return (
    <div
      className={ mergeTailwind(
        'row-span-1',
        'rounded-xl',
        'group/bento',
        'hover:shadow-xl',
        'transition duration-200',
        'shadow-input dark:shadow-none',
        'dark:bg-black bg-white ',
        'dark:border-white/[0.2] border border-transparent',
        'justify-between flex flex-col space-y-4 p-4',
        className
      ) }
    >
      {header}
      <div className='group-hover/bento:translate-x-2 transition duration-200'>
        { icon }
        <div className='font-sans font-bold text-neutral-600 dark:text-neutral-200 mb-2 mt-2'>
          { title }
        </div>
        <div className='font-sans font-normal text-neutral-600 text-xs dark:text-neutral-300'>
          { description }
        </div>
      </div>
    </div>
  );
};
