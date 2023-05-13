import { StoryObj, Meta } from '@storybook/react';
import SidebarView from './sidebar-new';

const meta: Meta<typeof SidebarView> = {
  title: 'Pages/Templates/Sidebar',
  component: SidebarView,
}

export default meta;

export const Default: StoryObj<typeof SidebarView> = {
  name: 'Default',
  args: {
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          allowForgotPassword: true,
          allowSignUp: true,
        }
      }
    ]
  }
}
