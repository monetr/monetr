/* eslint-disable no-console */
import React, { ReactElement, useEffect } from 'react';

import MLogo from '@monetr/interface/components/MLogo';
import MSpan from '@monetr/interface/components/MSpan';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';

import { TellerConnectEnrollment, TellerConnectFailure, useTellerConnect } from 'teller-connect-react';


export default function TellerSetup(): JSX.Element {
  const config = useAppConfiguration();
  const { open, ready, error } = useTellerConnect({
    applicationId: config.tellerApplicationId,
    environment: config.tellerEnvironment, // Hard coded for now.
    products: [
      'balance',
      'transactions',
    ],
    onSuccess: (result: TellerConnectEnrollment) => {
      console.log(result);
      return;
    },
    onEvent: (name: string, data: object) => {
      console.log({
        name,
        data,
      });
    },
    onExit: () => {
      console.log('teller exited!');
    },
    onFailure: (failure: TellerConnectFailure) => {
      console.warn(failure);
    },
    onInit: () => {
      console.log('teller inited');
    },
  });

  useEffect(() => {
    if (ready) {
      open();
    }
  }, [ready, open]);

  let inner: ReactElement = (
    <div className='flex flex-col justify-center items-center'>
      <MSpan className='text-2xl font-medium'>
        Getting Teller Ready!
      </MSpan>
      <MSpan className='text-lg' color='subtle'>
        One moment while we prepare your connection with Teller.
      </MSpan>
    </div>
  );

  if (error) {
    inner = (
      <div className='flex flex-col justify-center items-center'>
        <MSpan className='text-2xl font-medium'>
          Something isn't quite right
        </MSpan>
        <MSpan className='text-lg' color='subtle'>
          We were not able to get Teller ready for you at this time, please try again shortly.
        </MSpan>
        <MSpan className='text-lg' color='subtle'>
          If the problem continues, please contact support@monetr.app
        </MSpan>
        <MSpan className='text-md' color='muted'>
          Error Message: { error.message }
        </MSpan>
      </div>
    );
  }

  return (
    <div className='w-full h-full flex justify-center items-center gap-8 flex-col overflow-hidden text-center p-2'>
      <MLogo className='w-24 h-24' />
      { inner }
    </div>
  );
}
