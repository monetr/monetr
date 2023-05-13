import type { Preview } from "@storybook/react";
import { IconVariant, SnackbarProvider } from 'notistack';
import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import React from "react";
import { BrowserRouter as Router } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';

import '../ui/styles/styles.css'
import './preview.css';
import MockQueryClient, { MockRequest } from "./query";

export interface StoryArgs {
  requests: Array<MockRequest>;
}

const preview: Preview = {
  decorators: [
    (Story, context) => {
      const args = context.args as StoryArgs;
      const snackbarIcons: Partial<IconVariant> = {
        error: <ErrorIcon className="mr-2.5" />,
        success: <DoneIcon className="mr-2.5" />,
        warning: <WarningIcon className="mr-2.5" />,
        info: <InfoIcon className="mr-2.5" />,
      };
      return (
        <MockQueryClient requests={ args.requests || [] }>
          <Router>
            <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
              <NiceModal.Provider>
                <Story />
              </NiceModal.Provider>
            </SnackbarProvider>
          </Router>
        </MockQueryClient>
      );
    },
  ],
  args: {

  },
  parameters: {
    actions: { argTypesRegex: "^on[A-Z].*" },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/,
      },
    },
  },
};

export default preview;
