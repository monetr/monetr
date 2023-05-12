import React from 'react';
import MButton from './MButton';
import { StoryFn, Meta } from '@storybook/react';

export default {
  title: 'Button',
  component: MButton,
} as Meta<typeof MButton>;

export const Default: StoryFn<typeof MButton> = () => (
  <div className='w-full h-full flex'>
    <div className='max-w-5xl grid grid-cols-4 grid-flow-row gap-6'>
      <span className='w-full text-center'>Enabled</span>
      <span className='w-full text-center'>No Fill</span>
      <span className='w-full text-center'>Disabled</span>
      <span className='w-full text-center'>Disabled No Fill</span>
      <MButton theme='primary'>
        Login
      </MButton>
      <MButton theme='primary' kind='text'>
        Login
      </MButton>
      <MButton theme='primary' disabled>
        Login
      </MButton>
      <MButton theme='primary' disabled kind='text'>
        Login
      </MButton>

      <MButton theme='secondary'>
        Sign up
      </MButton>
      <MButton theme='secondary' kind='text'>
        Sign up
      </MButton>
      <MButton theme='secondary' disabled>
        Sign up
      </MButton>
      <MButton theme='secondary' disabled kind='text'>
        Sign up
      </MButton>

      <MButton theme='cancel'>
        Cancel
      </MButton>
      <MButton theme='cancel' kind='text'>
        Cancel
      </MButton>
      <MButton theme='cancel' disabled>
        Cancel
      </MButton>
      <MButton theme='cancel' disabled kind='text'>
        Cancel
      </MButton>
    </div>
  </div>
)


