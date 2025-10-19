import type { Meta, StoryFn } from '@storybook/react';
import { Plus } from 'lucide-react';
import { useSnackbar } from 'notistack';

import { Button } from './Button';
import FormButton from '@monetr/interface/components/FormButton';
import MForm from '@monetr/interface/components/MForm';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';

export default {
  title: '@monetr/interface/components/Button',
  component: Button,
} as Meta<typeof Button>;

export const Default: StoryFn<typeof Button> = () => (
  <div className='w-full flex p-4'>
    <div className='max-w-5xl grid grid-cols-6 grid-flow-row gap-6'>
      <span className='w-full text-center'>Enabled</span>
      <span className='w-full text-center'>Disabled</span>
      <span className='w-full text-center'>With Icon</span>
      <span className='w-full text-center'>With Icon Disabled</span>
      <span className='w-full text-center'>Icon Only</span>
      <span className='w-full text-center'>Icon Only Disabled</span>
      <Button variant='primary'>Primary</Button>
      <Button variant='primary' disabled>
        Primary
      </Button>
      <Button variant='primary'>
        <Plus />
        New Expense
      </Button>
      <Button variant='primary' disabled>
        <Plus />
        New Expense
      </Button>
      <Button variant='primary'>
        <Plus />
      </Button>
      <Button variant='primary' disabled>
        <Plus />
      </Button>

      <Button variant='secondary'>Secondary</Button>
      <Button variant='secondary' disabled>
        Secondary
      </Button>
      <Button variant='secondary'>
        <Plus />
        Secondary
      </Button>
      <Button variant='secondary' disabled>
        <Plus />
        Secondary
      </Button>
      <Button variant='secondary'>
        <Plus />
      </Button>
      <Button variant='secondary' disabled>
        <Plus />
      </Button>

      <Button variant='outlined'>Outlined</Button>
      <Button variant='outlined' disabled>
        Outlined
      </Button>
      <Button variant='outlined'>
        <Plus />
        Outlined
      </Button>
      <Button variant='outlined' disabled>
        <Plus />
        Outlined
      </Button>
      <Button variant='outlined'>
        <Plus />
      </Button>
      <Button variant='outlined' disabled>
        <Plus />
      </Button>

      <Button variant='text'>Text</Button>
      <Button variant='text' disabled>
        Text
      </Button>
      <Button variant='text'>
        <Plus />
        Text
      </Button>
      <Button variant='text' disabled>
        <Plus />
        Text
      </Button>
      <Button variant='text'>
        <Plus />
      </Button>
      <Button variant='text' disabled>
        <Plus />
      </Button>

      <Button variant='destructive'>Destructive</Button>
      <Button variant='destructive' disabled>
        Destructive
      </Button>
      <Button variant='destructive'>
        <Plus />
        Destructive
      </Button>
      <Button variant='destructive' disabled>
        <Plus />
        Destructive
      </Button>
      <Button variant='destructive'>
        <Plus />
      </Button>
      <Button variant='destructive' disabled>
        <Plus />
      </Button>
    </div>
  </div>
);

interface FormValues {
  name: string;
}

const initialValues: FormValues = {
  name: '',
};

export const Form: StoryFn<typeof Button> = () => {
  const { enqueueSnackbar } = useSnackbar();

  async function submit(values: FormValues) {
    enqueueSnackbar(`Form submitted: ${JSON.stringify(values)}`, {
      variant: 'success',
      disableWindowBlurListener: true,
    });
  }

  async function cancel() {
    enqueueSnackbar('Form canceled', {
      variant: 'warning',
      disableWindowBlurListener: true,
    });
  }

  return (
    <MForm initialValues={initialValues} onSubmit={submit} className='flex max-w-lg flex-col p-4'>
      <MSpan size='lg'>Hit enter should show submitted, not canceled</MSpan>
      <MTextField
        name='name'
        label='Name / Description'
        required
        autoComplete='off'
        placeholder='Amazon, Netflix...'
        data-1p-ignore
      />
      <div className='flex justify-end gap-2'>
        <Button variant='destructive' onClick={cancel}>
          Cancel
        </Button>
        <FormButton variant='primary' type='submit'>
          Create
        </FormButton>
      </div>
    </MForm>
  );
};
