import React from 'react';
import { Meta, StoryObj } from '@storybook/react';

import MDatePicker from './MDatePicker';
import MSpan from './MSpan';
import MTextField from './MTextField';

import { startOfTomorrow } from 'date-fns';


const meta: Meta<typeof MDatePicker> = {
  title: 'Components/Date Picker',
  component: MDatePicker,
};

export default meta;

export const Default: StoryObj<typeof MDatePicker> = {
  name: 'Default',
  render: () => (
    <div className="w-full flex p-4">
      <div className="w-full max-w-xl grid grid-cols-1 grid-flow-row gap-1">
        <MTextField 
          label="I'm a basic text field"
          labelDecorator={ () => <MSpan size='xs'>Just here for reference</MSpan> }
          placeholder='I can have some text here...' 
        />
        <MDatePicker
          label="When do you get paid next?"
          placeholder='Please select a date...'
          enableClear={ true }
        />
        <MDatePicker
          label="Must be in the future."
          placeholder='Please select a date...'
          min={ startOfTomorrow() }
        />
        <MDatePicker
          label="Go by year"
          placeholder='Please select a date...'
          enableYearNavigation
        />
        <MDatePicker
          label="Required date picker"
          placeholder='You must select a date...'
          required
        />
        <MDatePicker
          label="Disabled date picker"
          placeholder='You cannot select a date...'
          enableYearNavigation
          disabled
        />
        <MDatePicker
          label="With an error!"
          placeholder='Please select a date...'
          error='Invalid date selected!'
        />
      </div>
    </div>
  ),
};
