import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Check, ChevronsUpDown, CirclePlus, Settings } from 'lucide-react';

import { Button, buttonVariants } from '@monetr/interface/components/Button';
import { ComboboxItemProps, comboboxVariants } from '@monetr/interface/components/Combobox';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@monetr/interface/components/Command';
import { Drawer, DrawerContent, DrawerTrigger } from '@monetr/interface/components/Drawer';
import MBadge from '@monetr/interface/components/MBadge';
import MSpan from '@monetr/interface/components/MSpan';
import { Popover, PopoverContent, PopoverTrigger } from '@monetr/interface/components/Popover';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import { useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { useBankAccounts } from '@monetr/interface/hooks/useBankAccounts';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';
import { showNewBankAccountModal } from '@monetr/interface/modals/NewBankAccountModal';
import { BankAccountStatus } from '@monetr/interface/models/BankAccount';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';
import sortAccounts from '@monetr/interface/util/sortAccounts';

export default function SelectBankAccount(): JSX.Element {
  const navigate = useNavigate();
  const { data: allBankAccounts, isLoading: allIsLoading } = useBankAccounts();
  const { data: selectedBankAccount, isLoading: selectedIsLoading } = useSelectedBankAccount();
  const { data: link } = useCurrentLink();
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
      <div className='flex w-full gap-[1px]'>
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
                'h-[34px] group flex flex-auto'
              ) }
            >
              <div className='text-inherit flex-shrink truncate min-w-0'>
                { current?.value
                  ? accounts.find(option => option.value === current?.value)?.label
                  : 'Select a bank account...' }
              </div>
              <ChevronsUpDown
                className={ mergeTailwind('h-3 w-3 flex-none opacity-50 transition-opacity duration-100', {
                  'opacity-100': open,
                }) }
              />
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
        { link?.getIsManual() && (
          <Link
            role='combobox'
            aria-expanded={ open }
            className={ mergeTailwind(
              buttonVariants({ variant: 'text', size: 'select' }),
              comboboxVariants({ variant: 'text', size: 'select' }),
              'h-[34px] w-[34px] p-0 justify-center group rounded-tl-none rounded-bl-none shrink-0',
              'enabled:hover:ring-1',
              'enabled:hover:ring-dark-monetr-border-string',
              'focus:ring-dark-monetr-brand focus:ring-2',
            ) }
            to={ `/bank/${selectedBankAccount.bankAccountId}/settings` }
          >
            <Settings className='h-3 w-3 opacity-50 group-hover:opacity-100' />
          </Link>
        ) }
      </div>
    );
  }

  return (
    <div className='flex w-full gap-[1px]'>
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
              'h-[34px] group flex flex-auto'
            ) }
          >
            <div className='text-inherit flex-shrink truncate min-w-0'>
              { current?.value
                ? accounts.find(option => option.value === current?.value)?.label
                : 'Select a bank account...' }
            </div>
            <ChevronsUpDown
              className={ mergeTailwind('h-3 w-3 flex-none opacity-50 transition-opacity duration-100', {
                'opacity-100': open,
              }) }
            />
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
        { link?.getIsManual() && (
          <Link
            role='combobox'
            className={ mergeTailwind(
              buttonVariants({ variant: 'text', size: 'select' }),
              comboboxVariants({ variant: 'text', size: 'select' }),
              'h-[34px] w-[34px] p-0 justify-center group rounded-tl-none rounded-bl-none shrink-0',
              'enabled:hover:ring-1',
              'enabled:hover:ring-dark-monetr-border-string',
              'focus:ring-0', // DIFFERENT FROM MOBILE
            ) }
            to={ `/bank/${selectedBankAccount.bankAccountId}/settings` }
          >
            <Settings className='h-4 w-4 opacity-50 group-hover:opacity-100' />
          </Link>
        ) }
      </Popover>
    </div>
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
        { link?.getIsManual() && (
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
