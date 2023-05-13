import { Meta, StoryObj } from "@storybook/react";
import MLayout from ".";

const meta: Meta<typeof MLayout> = {
  title: 'Pages/Templates/Layout',
  component: MLayout,
};

export default meta;

export const Default: StoryObj<typeof MLayout> = {
  name: 'Default',
  args: {
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          billingEnabled: true,
        }
      }
    ]
  }
}
