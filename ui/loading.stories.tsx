import { Meta, StoryObj } from '@storybook/react';

import Loading from './loading';

const meta: Meta<typeof Loading> = {
  title: 'States',
  component: Loading,
};

export default meta;

export const Config: StoryObj<typeof Loading> = {
  name: 'Loading',
};

