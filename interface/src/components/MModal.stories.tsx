import React, { useState } from 'react';
import { Meta, StoryFn } from '@storybook/react';

import { MBaseButton } from './MButton';
import MModal from './MModal';

export default {
  title: 'Components/Modal',
  component: MModal,
} as Meta<typeof MModal>;


export const Default: StoryFn<typeof MModal> = () => {
  const [open, setOpen] = useState(false);

  return (
    <div>
      <MModal open={ open }>
        <h1>test!</h1>
        <MBaseButton onClick={ () => setOpen(false) }>
          Open!
        </MBaseButton>
      </MModal>
      <MBaseButton onClick={ () => setOpen(!open) }>
        Open!
      </MBaseButton>
    </div>
  );
};

