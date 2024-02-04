/* eslint-disable no-console */
import React, { ReactElement, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';

import MLogo from '@monetr/interface/components/MLogo';
import MSpan from '@monetr/interface/components/MSpan';
import MSpinner from '@monetr/interface/components/MSpinner';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import request from '@monetr/interface/util/request';

import { TellerConnectEnrollment, TellerConnectFailure, useTellerConnect } from 'teller-connect-react';


export default function TellerSetup(): JSX.Element {
  const config = useAppConfiguration();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  interface State {
    loading: boolean;
    settingUp: boolean;
    error: string | null;
    exited: boolean;
  }

  const [{ loading, error, exited, settingUp }, setState] = useState<Partial<State>>({
    error: null,
    exited: false,
    loading: false,
    settingUp: false,
  });

  async function longPollSetup(recur: number, linkId: number): Promise<void> {
    setState({
      loading,
      error,
      exited,
      settingUp: true,
    });

    if (recur > 6) {
      return Promise.resolve();
    }

    return void request()
      .get(`/teller/link/${ linkId }/wait`)
      .catch(error => {
        if (error.response.status === 408) {
          return longPollSetup(recur + 1, linkId);
        }

        throw error;
      });
  }

  const { open, ready } = useTellerConnect({
    applicationId: config.tellerApplicationId,
    environment: config.tellerEnvironment,
    // products: [
    //   'balance',
    //   'transactions',
    // ],
    onSuccess: async (result: TellerConnectEnrollment) => {
      setState({
        exited: true,
        settingUp: true,
      });
      return request().post('/teller/link', result)
        .then(async result => {
          const linkId: number = result.data.linkId;
          await longPollSetup(0, linkId);

          setTimeout(() => {
            queryClient.invalidateQueries(['/links']);
            queryClient.invalidateQueries(['/bank_accounts']);
            navigate('/');
          }, 8000);
        })
        .catch(error => {
          setState({
            error,
            loading,
            settingUp: false,
          });
        });
    },
    onEvent: (name: string, data: object) => {
      console.log({
        name,
        data,
      });
    },
    onExit: () => {
      setState({
        loading,
        exited: true,
      });
    },
    onFailure: (failure: TellerConnectFailure) => {
      console.warn(failure);
      setState({
        loading,
        exited,
        error: 'Teller exited with an error.',
      });
    },
    onInit: () => {
      console.log('teller inited');
    },
  });

  useEffect(() => {
    if (ready && !loading && !exited && !settingUp) {
      setTimeout(() => {
        open();
        setState({
          loading: true,
        });
      }, 1000);
    }
  }, [ready, open, loading, exited, settingUp]);

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

  if (loading) {
    inner = (
      <div className='flex flex-col justify-center items-center'>
        <MSpinner />
      </div>
    );
  }

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
          Error Message: { error }
        </MSpan>
      </div>
    );
  }

  if (settingUp) {
    inner = (
      <div className='flex flex-col justify-center items-center'>
        <MSpan className='text-2xl font-medium'>
          Successfully connected with Teller!
        </MSpan>
        <MSpan className='text-lg' color='subtle'>
          Hold on a moment while we pull the initial data from Teller into monetr.
        </MSpan>
      </div>
    );
  }


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
          Error Message: { error }
        </MSpan>
      </div>
    );
  }

  if (exited) {
    inner = (
      <div className='flex flex-col justify-center items-center'>
        <MSpan className='text-2xl font-medium'>
          Something isn't quite right
        </MSpan>
        <MSpan className='text-lg' color='subtle'>
          Teller exited, did you want to set it up later?
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
