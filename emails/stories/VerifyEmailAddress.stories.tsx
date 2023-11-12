import React from 'react';
import { render } from '@react-email/components';
import { Meta, StoryObj } from '@storybook/react';

// eslint-disable-next-line no-relative-import-paths/no-relative-import-paths
import VerifyEmailAddress from '../src/VerifyEmailAddress';

const meta: Meta<typeof VerifyEmailAddress> = {
  title: 'Email/Verify Email Address',
  component: VerifyEmailAddress,
};

export default meta;

export const Config: StoryObj<typeof VerifyEmailAddress> = {
  name: 'Default',
  render: () => {
    const html = render(
      <VerifyEmailAddress
        baseUrl='https://my.monetr.dev'
        firstName="Elliot"
        lastName="Courant"
        supportEmail="support@monetr.app"
        verifyLink='https://monetr.app'
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

