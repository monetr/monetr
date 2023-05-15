import { Meta, StoryObj } from "@storybook/react";
import ForgotPasswordPage from "./forgot-new";

const meta: Meta<typeof ForgotPasswordPage> = {
  title: 'Pages/Authentication/Forgot Password',
  component: ForgotPasswordPage,
};

export default meta;

export const Default: StoryObj<typeof ForgotPasswordPage> = {
  name: 'Default',
};
