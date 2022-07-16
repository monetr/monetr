import React from 'react';
import { Lock } from '@mui/icons-material';
import { Button } from '@mui/material';

export default function TOTPSettings(): JSX.Element {
  return (
    <div>
      <span className="text-2xl">
        Multi-Factor Authentication
      </span>
      <div className="w-full grid lg:grid-cols-2 gap-2.5 mt-2.5">
        <div className="flex items-center">
          <div className="w-full">
            <Button variant="contained" className="w-full">
              <Lock className="mr-2.5" />
              Setup TOTP
            </Button>
          </div>
        </div>
        <p className="opacity-70 h-full inline-block align-middle">
          You can add your MFA to an application like Google Authenticator or 1Password.
        </p>
      </div>
    </div>
  );
}
