import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Check, ChevronsUpDown, CirclePlus } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import { ComboboxItemProps, comboboxVariants } from '@monetr/interface/components/Combobox';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@monetr/interface/components/Command';
import { Drawer, DrawerContent, DrawerTrigger } from '@monetr/interface/components/Drawer';
import MBadge from '@monetr/interface/components/MBadge';
import MSpan from '@monetr/interface/components/MSpan';
import { Popover, PopoverContent, PopoverTrigger } from '@monetr/interface/components/Popover';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import { useBankAccounts, useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { useCurrentLink } from '@monetr/interface/hooks/links';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';
import { showNewBankAccountModal } from '@monetr/interface/modals/NewBankAccountModal';
import { BankAccountStatus } from '@monetr/interface/models/BankAccount';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';
import sortAccounts from '@monetr/interface/util/sortAccounts';

export default function SelectBankAccount(): JSX.Element {
  const navigate = useNavigate();
  const { data: allBankAccounts, isLoading: allIsLoading } = useBankAccounts();
  const { data: selectedBankAccount, isLoading: selectedIsLoading } = useSelectedBankAccount();
  const [open, setOpen] = React.useState(false);
  const isMobile = useIsMobile();

  const accounts: Array<SelectBankAccountPickerOption> = sortAccounts(allBankAccounts
    ?.filter(account => account.linkId === selectedBankAccount?.linkId))
    .map(account => ({
      label: account.name,
      value: account.bankAccountId,
      type: account.accountSubType,
      mask: account.mask,
      status: account.status,
    }));

  const current = accounts?.find(account => account.value === selectedBankAccount?.bankAccountId);

  if (allIsLoading || selectedIsLoading) {
    return (
      <Skeleton className='w-full h-10' />
    );
  }

  if (isMobile) {
    return (
      <Drawer open={ open } onOpenChange={ setOpen }>
        <DrawerTrigger asChild>
          <Button
            size='select'
            variant='text'
            role='combobox'
            aria-expanded={ open }
            disabled={ false }
            className={ mergeTailwind(
              comboboxVariants({ variant: 'text', size: 'select' }),
              'w-full h-[34px] test'
            ) }
          >
            <div className='text-ellipsis text-nowrap min-w-0 overflow-hidden text-inherit'>
              { current?.value
                ? accounts.find(option => option.value === current?.value)?.label
                : 'Select a bank account...' }
            </div>
            <ChevronsUpDown className='h-3 w-3 shrink-0 opacity-50' />
          </Button>
        </DrawerTrigger>
        <DrawerContent>
          <SelectBankAccountPicker
            value={ current?.value }
            setOpen={ setOpen }
            options={ accounts }
            onSelect={ value => navigate(`/bank/${value}/transactions`) }
          />
        </DrawerContent>
      </Drawer>
    );
  }

  return (
    <Popover open={ open } onOpenChange={ setOpen }>
      <PopoverTrigger asChild>
        <Button
          size='select'
          variant='text'
          role='combobox'
          aria-expanded={ open }
          disabled={ false }
          className={ mergeTailwind(
            comboboxVariants({ variant: 'text', size: 'select' }), 
            'w-full h-[34px] test'
          ) }
        >
          <div className='text-ellipsis text-nowrap min-w-0 overflow-hidden text-inherit'>
            { current?.value
              ? accounts.find(option => option.value === current?.value)?.label
              : 'Select a bank account...' }
          </div>
          <ChevronsUpDown className='h-3 w-3 shrink-0 opacity-50' />
        </Button>
      </PopoverTrigger>
      <PopoverContent className='w-80'>
        <SelectBankAccountPicker
          value={ current?.value }
          setOpen={ setOpen }
          options={ accounts }
          onSelect={ value => navigate(`/bank/${value}/transactions`) }
        />
      </PopoverContent>
    </Popover>
  );
}

interface SelectBankAccountPickerOption {
  value: string;
  label: string;
  disabled?: boolean;
  // Fields used for rich display.
  type: string;
  mask: string;
  status: BankAccountStatus;
}

function BankAccountSelectItem(props: ComboboxItemProps<string, SelectBankAccountPickerOption>): JSX.Element {
  return (
    <div className='flex items-center w-full gap-1'>
      <Check
        className={ mergeTailwind(
          'mr-1 h-5 w-5 flex-none',
          props.currentValue == props.option.value ? 'opacity-100' : 'opacity-0'
        ) }
      />
      <MSpan className='w-full' color='emphasis' ellipsis>
        { props.option.label }
      </MSpan>
      { props.option.status === 'inactive' && (
        <MBadge size='xs' className='bg-dark-monetr-border-subtle'>
          Inactive
        </MBadge>
      ) }
      { props.option.mask != '' && (
        <MBadge size='xs' className='font-mono'>
          { props.option.mask }
        </MBadge>
      ) }
    </div>
  );
}

interface SelectBankAccountPickerProps {
  searchPlaceholder?: string;
  emptyString?: string;
  options: Array<SelectBankAccountPickerOption>;
  value?: string;
  onSelect?: (value: string) => void;
  setOpen: (open: boolean) => void;
}

function SelectBankAccountPicker(props: SelectBankAccountPickerProps): JSX.Element {
  const isMobile = useIsMobile();
  const { data: link } = useCurrentLink();
  return (
    <Command>
      { !isMobile && <CommandInput placeholder={ props.searchPlaceholder } /> }
      <CommandList>
        <CommandEmpty>{ props.emptyString }</CommandEmpty>
        { link.getIsManual() && (
          <CommandGroup className='' heading='Controls'>
            <CommandItem
              value='null'
              title='Create an account'
              onSelect={ () => {
                props.setOpen(false);
                showNewBankAccountModal();
              } }
            >
              <div className='flex items-center w-full gap-1'>
                <CirclePlus className='mr-1 h-5 w-5 flex-none' />
                <MSpan className='w-full' color='emphasis' ellipsis>
                  Add Another Account
                </MSpan>
              </div>
            </CommandItem>
          </CommandGroup>
        ) }
        <CommandGroup className='pb-6 md:pb-1' heading='Accounts'>
          { props.options.map(option => (
            <CommandItem
              key={ option.value }
              value={ `${option.label} ${option.value}` /* makes search work properly :( */ }
              title={ option.label }
              onSelect={ () => {
                props.onSelect && props.onSelect(option.value);
                props.setOpen(false);
              } }
            >
              <BankAccountSelectItem currentValue={ props.value } option={ option } />
            </CommandItem>
          )) }
        </CommandGroup>
      </CommandList>
    </Command>
  );
}
