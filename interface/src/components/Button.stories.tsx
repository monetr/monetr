import React from 'react';
import { AddOutlined } from '@mui/icons-material';
import { Meta, StoryFn } from '@storybook/react';

import { Button } from './Button';

export default {
  title: '@monetr/interface/components/Button',
  component: Button,
} as Meta<typeof Button>;

export const Default: StoryFn<typeof Button> = () => (
  <div className='w-full flex p-4'>
    <div className='max-w-5xl grid grid-cols-6 grid-flow-row gap-6'>
      <span className='w-full text-center'>Enabled</span>
      <span className='w-full text-center'>Disabled</span>
      <span className='w-full text-center'>With Icon</span>
      <span className='w-full text-center'>With Icon Disabled</span>
      <span className='w-full text-center'>Icon Only</span>
      <span className='w-full text-center'>Icon Only Disabled</span>
      <Button variant='primary'>
        Primary
      </Button>
      <Button variant='primary' disabled>
        Primary
      </Button>
      <Button variant='primary'>
        <AddOutlined />
        New Expense
      </Button>
      <Button variant='primary' disabled>
        <AddOutlined />
        New Expense
      </Button>
      <Button variant='primary'>
        <AddOutlined />
      </Button>
      <Button variant='primary' disabled>
        <AddOutlined />
      </Button>

      <Button variant='secondary'>
        Secondary
      </Button>
      <Button variant='secondary' disabled>
        Secondary
      </Button>
      <Button variant='secondary'>
        <AddOutlined />
        Secondary
      </Button>
      <Button variant='secondary' disabled>
        <AddOutlined />
        Secondary
      </Button>
      <Button variant='secondary'>
        <AddOutlined />
      </Button>
      <Button variant='secondary' disabled>
        <AddOutlined />
      </Button>

      <Button variant='outlined'>
        Outlined
      </Button>
      <Button variant='outlined' disabled>
        Outlined
      </Button>
      <Button variant='outlined'>
        <AddOutlined />
        Outlined
      </Button>
      <Button variant='outlined' disabled>
        <AddOutlined />
        Outlined
      </Button>
      <Button variant='outlined'>
        <AddOutlined />
      </Button>
      <Button variant='outlined' disabled>
        <AddOutlined />
      </Button>

      <Button variant='text'>
        Text
      </Button>
      <Button variant='text' disabled>
        Text
      </Button>
      <Button variant='text'>
        <AddOutlined />
        Text
      </Button>
      <Button variant='text' disabled>
        <AddOutlined />
        Text
      </Button>
      <Button variant='text'>
        <AddOutlined />
      </Button>
      <Button variant='text' disabled>
        <AddOutlined />
      </Button>

      <Button variant='destructive'>
        Destructive
      </Button>
      <Button  variant='destructive' disabled>
        Destructive
      </Button>
      <Button  variant='destructive'>
        <AddOutlined />
        Destructive
      </Button>
      <Button  variant='destructive' disabled>
        <AddOutlined />
        Destructive
      </Button>
      <Button  variant='destructive'>
        <AddOutlined />
      </Button>
      <Button  variant='destructive' disabled>
        <AddOutlined />
      </Button>
    </div>
  </div>
);

