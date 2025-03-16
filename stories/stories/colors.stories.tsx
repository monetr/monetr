import React, { Component } from 'react';
import { Meta, StoryObj } from '@storybook/react';
import resolveConfig from 'tailwindcss/resolveConfig';

// eslint-disable-next-line no-relative-import-paths/no-relative-import-paths
import tailwindConfig from '../tailwind.config.ts';

const realTailwindConfig = resolveConfig(tailwindConfig);

const meta: Meta<typeof Component> = {
  title: 'Colors',
  parameters: {
  },
};

export default meta;

export const Pallete: StoryObj<typeof Component> = {
  name: 'Pallete',
  render: () => {
    const theme = realTailwindConfig.theme;

    const colors = Object.entries(theme.colors).map(([name, variations]) => {
      const items = Object.entries(variations).filter(([_, color]) => !!color).map(([weight, color]) => (
        <div key={ `${name}-${weight}` }>
          <span>{ name }-{ weight }</span>
          <div className='p-4 col-span-1' style={ { backgroundColor: color as any } } />
        </div>
      ));

      return items;
    }).flatMap(item => item);

    return (
      <div className='bg-black'>
        <div className='grid grid-cols-10 gap-y-4 gap-x-2'>
          { colors }
        </div>
      </div>
    );
  },
};
