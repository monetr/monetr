import React from 'react';
import { Meta, StoryFn } from '@storybook/react';

import MLink from './MLink';

export default {
  title: 'Components/Link',
  component: MLink,
} as Meta<typeof MLink>;

export const Default: StoryFn<typeof MLink> = () => (
  <div className="w-full flex p-4">
    <MLink to="#">I am a link!</MLink>
  </div>
);
