import { Meta, StoryFn } from "@storybook/react";
import React from "react";
import MLink from "./MLink";

export default {
  title: 'Components/Link',
  component: MLink,
} as Meta<typeof MLink>;

export const Default: StoryFn<typeof MLink> = () => (
  <MLink to="#">I am a link!</MLink>
);
