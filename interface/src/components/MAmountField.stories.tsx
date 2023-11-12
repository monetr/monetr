import React from 'react';
import { Meta, StoryFn } from '@storybook/react';

import MAmountField from './MAmountField';


export default {
  title: '@monetr/interface/components/Amount Field',
  component: MAmountField,
} as Meta<typeof MAmountField>;

export const Default: StoryFn<typeof MAmountField> = () => (
  <div className='w-full flex p-4'>
    <div className='w-full max-w-xl grid grid-cols-1 grid-flow-row gap-1'>
      <MAmountField label='Amount' name='name' value={ 0 } />
      <MAmountField label='Amount Non-Negative' name='name' value={ 0 } allowNegative={ false } />
      <MAmountField label='Amount (Required)' name='name' required />
      <MAmountField label='Amount (Disabled)' name='name' disabled  value={ 0 } />
      <MAmountField label='Amount (Disabled)' name='name' error='Not a valid amount!' />
    </div>
  </div>
);
