import React from 'react';
import { Meta, StoryFn } from '@storybook/react';

import { Calendar } from './Calendar';

export default {
  title: '@monetr/interface/components/Calendar',
  component: Calendar,
} as Meta<typeof Calendar>;

export const Default: StoryFn<typeof Calendar> = () => {
  const [date, setDate] = React.useState<Date | undefined>(new Date());

  return (
    <div className='w-full flex p-4'>
      <div className='max-w-5xl grid grid-cols-2 grid-flow-row gap-6'>
        <span className='w-full text-center'>Enabled</span>
        <span className='w-full text-center'>Disabled</span>
        <Calendar
          mode='single'
          selected={date}
          onSelect={setDate}
          className='rounded-md border border-dark-monetr-border'
        />
        <Calendar
          mode='single'
          selected={date}
          onSelect={setDate}
          disabled
          className='rounded-md border border-dark-monetr-border'
        />
      </div>
    </div>
  );
};
