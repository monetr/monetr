import { Meta, StoryObj } from '@storybook/react';

import ConfigError from './config';

const meta: Meta<typeof ConfigError> = {
  title: 'Errors',
  component: ConfigError,
};

export default meta;

export const Config: StoryObj<typeof ConfigError> = {
  name: 'Config',
};

