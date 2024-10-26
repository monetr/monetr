import React, { Component } from 'react';
import { Meta, StoryObj } from '@storybook/react';

import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@monetr/interface/components/Dialog';
import MButton from '@monetr/interface/components/MButton';
import MTextField from '@monetr/interface/components/MTextField';


const meta: Meta<typeof Component> = {
  title: 'Radix UI',
  parameters: {
    // msw: {
    //   handlers: [
    //     // ...GetAPIFixtures(),
    //   ],
    // },
  },
};

export default meta;

export const Transactions: StoryObj<typeof Component> = {
  name: 'Radix Playground',
  render: () => (
    <Dialog>
      <DialogTrigger asChild>
        <MButton variant='outlined'>Edit Profile</MButton>
      </DialogTrigger>
      <DialogContent className='sm:max-w-[425px]'>
        <DialogHeader>
          <DialogTitle>Edit profile</DialogTitle>
          <DialogDescription>
            Make changes to your profile here. Click save when you're done.
          </DialogDescription>
        </DialogHeader>
        <div className='grid gap-4 py-4'>
          <div className='grid grid-cols-4 items-center gap-4'>
            <MTextField
              label='Name'
              name='name'
              className='col-span-3'
            />
          </div>
          <div className='grid grid-cols-4 items-center gap-4'>
            <MTextField
              label='Username'
              name='username'
              className='col-span-3'
            />
          </div>
        </div>
        <DialogFooter>
          <MButton type='submit'>Save changes</MButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
  // parameters: {
  //   reactRouter: {
  //     routePath: '/*',
  //     browserPath: '/bank/12/transactions',
  //   },
  // },
};
