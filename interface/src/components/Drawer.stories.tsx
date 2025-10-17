/* eslint-disable max-len */

import type { Meta, StoryFn } from '@storybook/react';

import { Button } from '@monetr/interface/components/Button';
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
  DrawerWrapper,
} from '@monetr/interface/components/Drawer';

export default {
  title: '@monetr/interface/components/Drawer',
  component: Drawer,
} as Meta<typeof Drawer>;

export const Default: StoryFn<typeof Drawer> = () => (
  <div className='p-4'>
    <Drawer>
      <DrawerTrigger asChild>
        <Button variant='outlined'>Open Drawer</Button>
      </DrawerTrigger>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle>Are you absolutely sure?</DrawerTitle>
          <DrawerDescription>This action cannot be undone.</DrawerDescription>
        </DrawerHeader>
        <DrawerWrapper>
          <div className='h-[200px] align-middle text-center'>Big space</div>
        </DrawerWrapper>
        <DrawerFooter>
          <Button>Submit</Button>
          <DrawerClose>
            <Button className='w-full' variant='outlined'>
              Cancel
            </Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  </div>
);
