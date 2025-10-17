import React, { useState } from 'react';
import { Meta, StoryFn } from '@storybook/react';

import MModal from './MModal';
import { Button } from '@monetr/interface/components/Button';

export default {
  title: '@monetr/interface/components/Modal',
  component: MModal,
} as Meta<typeof MModal>;

export const Default: StoryFn<typeof MModal> = () => {
  const [open, setOpen] = useState(false);

  return (
    <div>
      <MModal open={open}>
        <h1>test!</h1>
        <Button onClick={() => setOpen(false)}>Open!</Button>
      </MModal>
      <Button onClick={() => setOpen(!open)}>Open!</Button>
    </div>
  );
};
