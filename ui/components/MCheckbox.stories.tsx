import { Checkbox, FormControlLabel } from "@mui/material";
import { Meta, StoryFn } from "@storybook/react";
import React from "react";
import MCheckbox from "./MCheckbox";

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
        <MCheckbox id='test' label='Remember me for 30 days' />
        <div>
          <FormControlLabel control={<Checkbox defaultChecked />} label="Remember me for 30 days" />
        </div>
      </div>
    </fieldset>
  </div>
)
