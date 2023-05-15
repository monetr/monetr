import React from 'react';
import { Meta, StoryFn } from '@storybook/react';

import MButton from './MButton';

export default {
  title: 'Components/Button',
  component: MButton,
} as Meta<typeof MButton>;

export const Default: StoryFn<typeof MButton> = () => (
  <div className='w-full flex p-4'>
    <div className='max-w-5xl grid grid-cols-4 grid-flow-row gap-6'>
      <span className='w-full text-center'>Enabled</span>
      <span className='w-full text-center'>No Fill</span>
      <span className='w-full text-center'>Disabled</span>
      <span className='w-full text-center'>Disabled No Fill</span>
      <MButton color='primary'>
        Login
      </MButton>
      <MButton color='primary' variant='text'>
        Login
      </MButton>
      <MButton color='primary' disabled>
        Login
      </MButton>
      <MButton color='primary' disabled variant='text'>
        Login
      </MButton>

      <MButton color='secondary'>
        Sign up
      </MButton>
      <MButton color='secondary' variant='text'>
        Sign up
      </MButton>
      <MButton color='secondary' disabled>
        Sign up
      </MButton>
      <MButton color='secondary' disabled variant='text'>
        Sign up
      </MButton>

      <MButton color='cancel'>
        Cancel
      </MButton>
      <MButton color='cancel' variant='text'>
        Cancel
      </MButton>
      <MButton color='cancel' disabled>
        Cancel
      </MButton>
      <MButton color='cancel' disabled variant='text'>
        Cancel
      </MButton>
    </div>
  </div>
);


