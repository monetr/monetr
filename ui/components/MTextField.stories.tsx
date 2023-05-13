import { Divider } from "@mui/material";
import { Meta, StoryFn } from "@storybook/react";
import React from "react";
import MTextField from "./MTextField";


export default {
  title: 'Text Field',
  component: MTextField,
} as Meta<typeof MTextField>

export const Default: StoryFn<typeof MTextField> = () => (
  <div className="w-full flex">
    <div className="w-full max-w-xl grid grid-cols-1 grid-flow-row gap-6">
      <MTextField label="Name" type='text' name='name' />
      <MTextField label="Email Address" type='email' name='email' />
      <MTextField label="Password" type='password' name='password' />
      <MTextField label="Amount" type='number' name='amount' />
      <Divider />
      <MTextField label="Name" type='text' name='name' disabled />
    </div>
  </div>
);
