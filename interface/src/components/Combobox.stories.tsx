import type { Meta, StoryFn } from '@storybook/react';

import { Combobox } from './Combobox';

export default {
  title: '@monetr/interface/components/Combobox',
  component: Combobox,
} as Meta<typeof Combobox>;

const frameworks = [
  {
    value: 'next.js',
    label: 'Next.js',
  },
  {
    value: 'sveltekit',
    label: 'SvelteKit',
  },
  {
    value: 'nuxt.js',
    label: 'Nuxt.js',
  },
  {
    value: 'remix',
    label: 'Remix',
  },
  {
    value: 'astro',
    label: 'Astro',
  },
];

export const Default: StoryFn<typeof Combobox> = () => (
  <div className='w-full flex p-4'>
    <div className='max-w-5xl grid grid-cols-2 gap-6'>
      <span className='w-full text-center'>Enabled</span>
      <span className='w-full text-center'>Disabled</span>
      <Combobox size='default' variant='outlined' placeholder='Outlined Placeholder...' options={frameworks} />
      <Combobox disabled size='default' variant='outlined' placeholder='Outlined Placeholder...' options={frameworks} />

      <Combobox size='md' variant='outlined' placeholder='Outlined Placeholder...' options={frameworks} />
      <Combobox disabled size='md' variant='outlined' placeholder='Outlined Placeholder...' options={frameworks} />

      <Combobox size='default' variant='text' placeholder='Placeholder...' options={frameworks} />
      <Combobox disabled size='default' variant='text' placeholder='Placeholder...' options={frameworks} />

      <Combobox size='md' variant='text' placeholder='Text Placeholder...' options={frameworks} />
      <Combobox disabled size='md' variant='text' placeholder='Text Placeholder...' options={frameworks} />
    </div>
  </div>
);
