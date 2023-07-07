/* eslint-disable max-len */
import React from 'react';
import { Link, Route, Routes } from 'react-router-dom';

import MDivider from 'components/MDivider';
import MSpan from 'components/MSpan';
import SettingsOverview from 'pages/settings/overview';
import SettingsSecurity from 'pages/settings/security';

export default function Settings(): JSX.Element {
  return (
    <div className='flex flex-col w-full py-4 h-full relative'>
      <MSpan className='mx-4 text-5xl font-medium'>
        Settings
      </MSpan>
      <div className='w-full flex px-4 mt-4 gap-6'>
        <Link to="/settings">
          <MSpan className='cursor-pointer dark:hover:text-dark-monetr-content-emphasis font-bold dark:text-dark-monetr-brand-faint'>
            Overview
          </MSpan>
        </Link>
        <Link to="/settings/security">
          <MSpan className='cursor-pointer dark:hover:text-dark-monetr-content-emphasis font-normal'>
            Security
          </MSpan>
        </Link>
        <MSpan className='cursor-pointer dark:hover:text-dark-monetr-content-emphasis font-normal'>
          About
        </MSpan>
      </div>
      <MDivider className='mt-3' />
      <div className='w-full h-full'>
        <Routes>
          <Route index path='/settings' element={ <SettingsOverview /> } />
          <Route path='/settings/security' element={ <SettingsSecurity /> } />
        </Routes>
      </div>
    </div>
  );
}
