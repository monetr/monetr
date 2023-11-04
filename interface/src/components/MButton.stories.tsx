import React from 'react';
import { Meta, StoryFn } from '@storybook/react';

import MFormButton from './MButton';

export default {
  title: '@monetr/interface/components/Button',
  component: MFormButton,
} as Meta<typeof MFormButton>;

export const Default: StoryFn<typeof MFormButton> = () => (
  <div className='w-full flex p-4'>
    <div className='max-w-5xl grid grid-cols-4 grid-flow-row gap-6'>
      <span className='w-full text-center'>Enabled</span>
      <span className='w-full text-center'>No Fill</span>
      <span className='w-full text-center'>Disabled</span>
      <span className='w-full text-center'>Disabled No Fill</span>
      <MFormButton color='primary'>
        Login
      </MFormButton>
      <MFormButton color='primary' variant='text'>
        Login
      </MFormButton>
      <MFormButton color='primary' disabled>
        Login
      </MFormButton>
      <MFormButton color='primary' disabled variant='text'>
        Login
      </MFormButton>

      <MFormButton color='secondary'>
        Sign up
      </MFormButton>
      <MFormButton color='secondary' variant='text'>
        Sign up
      </MFormButton>
      <MFormButton color='secondary' disabled>
        Sign up
      </MFormButton>
      <MFormButton color='secondary' disabled variant='text'>
        Sign up
      </MFormButton>

      <MFormButton color='cancel'>
        Cancel
      </MFormButton>
      <MFormButton color='cancel' variant='text'>
        Cancel
      </MFormButton>
      <MFormButton color='cancel' disabled>
        Cancel
      </MFormButton>
      <MFormButton color='cancel' disabled variant='text'>
        Cancel
      </MFormButton>
    </div>
  </div>
);


