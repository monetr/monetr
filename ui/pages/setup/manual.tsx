import React from 'react';

import MForm from 'components/MForm';
import MLink from 'components/MLink';
import MLogo from 'components/MLogo';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';

interface SetupManualValues {
  name: string;
  initialBalance: number;
}

const initialValues: SetupManualValues = {
  name: '',
  initialBalance: 0,
};

export default function SetupManual(): JSX.Element {

  return (
    <div className='w-full h-full flex justify-center items-center gap-8 flex-col overflow-hidden text-center p-2'>
      <MLogo className='w-24 h-24' />
      <div className='flex flex-col justify-center items-center text-center'>
        <MSpan size='2xl' weight='medium'>
            Welcome to monetr!
        </MSpan>
        <MSpan size='lg' color='subtle'>
          Let's create a new budget to get started. What do you want to call this budget?
        </MSpan>
        <MForm
          initialValues={ initialValues }
          onSubmit={ () => {} }
        >
          <MTextField 
            name='name'
            label='Budget Name'
          />
        </MForm>
      </div>
      <LogoutFooter />
    </div>
  );
}

function LogoutFooter(): JSX.Element {
  return (
    <div className='flex justify-center gap-1'>
      <MSpan color="subtle" size='sm'>Not ready to continue?</MSpan>
      <MLink to="/logout" size="sm">Logout for now</MLink>
    </div>
  );
}
