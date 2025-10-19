import { useState } from 'react';
import type { Meta, StoryFn } from '@storybook/react';

import { Button } from '@monetr/interface/components/Button';

import MModal from './MModal';

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
