import React, { Fragment } from 'react';
import { Meta, StoryFn } from '@storybook/react';

import MCheckbox from './MCheckbox';

export default {
  title: 'Components/Checkbox',
  component: MCheckbox,
} as Meta<typeof MCheckbox>;

export const Default: StoryFn<typeof MCheckbox> = () => (
  <div className="w-full flex p-4">
    <fieldset>
      <div className="max-w-5xl grid grid-cols-2 grid-flow-row gap-6">
        <span className='w-full text-center'>Enabled</span>
        <span className='w-full text-center'>Disabled</span>
        <MCheckbox
          id='test'
          label='Remember me for 30 days'
        />
        <MCheckbox
          id='test2'
          label='Remember me for 30 days'
          disabled
        />
        <MCheckbox
          id='test0'
          label='Remember me for 30 days'
          checked
        />
        <MCheckbox
          id='test-1'
          label='Remember me for 30 days'
          disabled
          checked
        />
        <MCheckbox
          id='test3'
          label='Remember me for 30 days'
          description="Keep yourself logged in for a while."
        />
        <MCheckbox
          id='test4'
          label='Remember me for 30 days'
          description="Keep yourself logged in for a while."
          disabled
        />
        <MCheckbox
          id='test5'
          label='Remember me for 30 days'
          description="I am a much longer description, I want to see how these will work when descriptions are just very long. What if the description has a lot to say."
        />
        <MCheckbox
          id='test5'
          label='Remember me for 30 days'
          description="I am a much longer description, I want to see how these will work when descriptions are just very long. What if the description has a lot to say."
          disabled
        />
        <MCheckbox
          id='test6'
          label={
            <Fragment>
              I agree to monetr's&nbsp;
              <a
                target="_blank"
                className="text-blue-500 hover:underline focus:ring-2 focus:ring-blue-500 focus:underline"
                href='https://github.com/monetr/legal/blob/main/TERMS_OF_USE.md'>
                Terms of Use
              </a> and&nbsp;
              <a
                target="_blank"
                className="text-blue-500 hover:underline focus:ring-2 focus:ring-blue-500 focus:underline"
                href='https://github.com/monetr/legal/blob/main/PRIVACY.md'
              >
                Privacy Policy
              </a>
            </Fragment>
          }
        />
        <MCheckbox
          id='test7'
          label={
            <Fragment>
              I agree to monetr's&nbsp;
              <a
                target="_blank"
                className="text-blue-500 hover:underline focus:ring-2 focus:ring-blue-500 focus:underline"
                href='https://github.com/monetr/legal/blob/main/TERMS_OF_USE.md'>
                Terms of Use
              </a> and&nbsp;
              <a
                target="_blank"
                className="text-blue-500 hover:underline focus:ring-2 focus:ring-blue-500 focus:underline"
                href='https://github.com/monetr/legal/blob/main/PRIVACY.md'
              >
                Privacy Policy
              </a>
            </Fragment>
          }
          disabled
        />
      </div>
    </fieldset>
  </div>
);
