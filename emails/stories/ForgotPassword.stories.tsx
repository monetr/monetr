import React from 'react';
import { render } from '@react-email/components';
import { Meta, StoryObj } from '@storybook/react';

// eslint-disable-next-line no-relative-import-paths/no-relative-import-paths
import ForgotPassword from '../src/ForgotPassword';

const meta: Meta<typeof ForgotPassword> = {
  title: 'Email/Forgot Password',
  component: ForgotPassword,
};

export default meta;

export const Config: StoryObj<typeof ForgotPassword> = {
  name: 'Default',
  render: () => {
    const html = render(
      <ForgotPassword
        baseUrl='https://my.monetr.dev'
        firstName="Elliot"
        lastName="Courant"
        supportEmail="support@monetr.app"
        resetUrl='https://monetr.app'
      />,
      {
        pretty: true,
      },
    );

    return (
      <div
        className='absolute top-0 bg-white h-full w-full email-preview'
        dangerouslySetInnerHTML={ { __html: html } }
      />
    );
  },
};

