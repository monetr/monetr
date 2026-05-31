import React from 'react';
import { Check, ChevronsUpDown, CirclePlus, Settings } from 'lucide-react';
import { Link, useLocation } from 'wouter';

import Badge from '@monetr/interface/components/Badge';
import { Button, buttonVariants } from '@monetr/interface/components/Button';
import { type ComboboxItemProps, comboboxVariants } from '@monetr/interface/components/Combobox';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@monetr/interface/components/Command';
import { Drawer, DrawerContent, DrawerTrigger } from '@monetr/interface/components/Drawer';
import Flex from '@monetr/interface/components/Flex';
import { layoutVariants } from '@monetr/interface/components/Layout';
import { Popover, PopoverContent, PopoverTrigger } from '@monetr/interface/components/Popover';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import Typography from '@monetr/interface/components/Typography';
import { useBankAccounts } from '@monetr/interface/hooks/useBankAccounts';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { showNewBankAccountModal } from '@monetr/interface/modals/NewBankAccountModal';
import type { BankAccountStatus } from '@monetr/interface/models/BankAccount';
import mergeClasses from '@monetr/interface/util/mergeClasses';
import sortAccounts from '@monetr/interface/util/sortAccounts';

import styles from './SelectBankAccount.module.scss';

export default function SelectBankAccount(): JSX.Element {
  const [, navigate] = useLocation();
  const { data: allBankAccounts, isLoading: allIsLoading } = useBankAccounts();
  const { data: selectedBankAccount, isLoading: selectedIsLoading } = useSelectedBankAccount();
  const [open, setOpen] = React.useState(false);
  const isMobile = useIsMobile();

  const accounts: Array<SelectBankAccountPickerOption> = sortAccounts(
    allBankAccounts?.filter(account => account.linkId === selectedBankAccount?.linkId),
  ).map(account => ({
    label: account.name,
    value: account.bankAccountId,
    type: account.accountSubType,
    mask: account.mask,
    status: account.status,
  }));

  const current = accounts?.find(account => account.value === selectedBankAccount?.bankAccountId);

  if (allIsLoading || selectedIsLoading) {
    return <Skeleton className={styles.skeleton} />;
  }

  if (isMobile) {
    return (
      <div className={styles.row}>
        <Drawer onOpenChange={setOpen} open={open}>
          <DrawerTrigger asChild>
            <Button
              aria-expanded={open}
              className={mergeClasses(comboboxVariants({ variant: 'text', size: 'select' }), styles.trigger)}
              disabled={false}
              role='combobox'
              size='select'
              variant='text'
            >
              <div className={styles.triggerLabel}>
                {current?.value
                  ? accounts.find(option => option.value === current?.value)?.label
                  : 'Select a bank account...'}
              </div>
              <ChevronsUpDown className={mergeClasses(styles.chevron, { [styles.chevronOpen]: open })} />
            </Button>
          </DrawerTrigger>
          <DrawerContent>
            <SelectBankAccountPicker
              onSelect={value => navigate(`/bank/${value}/transactions`)}
              options={accounts}
              setOpen={setOpen}
              value={current?.value}
            />
          </DrawerContent>
        </Drawer>
        <Link
          aria-expanded={open}
          className={mergeClasses(
            buttonVariants({ variant: 'text', size: 'select' }),
            comboboxVariants({ variant: 'text', size: 'select' }),
            styles.settingsLinkMobile,
          )}
          role='combobox'
          to={`/bank/${selectedBankAccount?.bankAccountId}/settings`}
        >
          <Settings className={styles.settingsIconMobile} />
        </Link>
      </div>
    );
  }

  return (
    <div className={styles.row}>
      <Popover onOpenChange={setOpen} open={open}>
        <PopoverTrigger asChild>
          <Button
            aria-expanded={open}
            className={mergeClasses(comboboxVariants({ variant: 'text', size: 'select' }), styles.trigger)}
            disabled={false}
            role='combobox'
            size='select'
            variant='text'
          >
            <div className={styles.triggerLabel}>
              {current?.value
                ? accounts.find(option => option.value === current?.value)?.label
                : 'Select a bank account...'}
            </div>
            <ChevronsUpDown className={mergeClasses(styles.chevron, { [styles.chevronOpen]: open })} />
          </Button>
        </PopoverTrigger>
        <PopoverContent className={styles.popover}>
          <SelectBankAccountPicker
            onSelect={value => navigate(`/bank/${value}/transactions`)}
            options={accounts}
            setOpen={setOpen}
            value={current?.value}
          />
        </PopoverContent>
        <Link
          className={mergeClasses(
            buttonVariants({ variant: 'text', size: 'select' }),
            comboboxVariants({ variant: 'text', size: 'select' }),
            styles.settingsLinkDesktop, // No focus ring, different from mobile.
          )}
          role='combobox'
          to={`/bank/${selectedBankAccount?.bankAccountId}/settings`}
        >
          <Settings className={styles.settingsIconDesktop} />
        </Link>
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
  mask?: string;
  status: BankAccountStatus;
}

function BankAccountSelectItem(props: ComboboxItemProps<string, SelectBankAccountPickerOption>): JSX.Element {
  return (
    <Flex align='center' gap='sm'>
      <Check
        className={mergeClasses(
          styles.check,
          props.currentValue === props.option.value ? styles.checkVisible : styles.checkHidden,
        )}
      />
      <Typography className={layoutVariants({ width: 'full' })} color='emphasis' ellipsis>
        {props.option.label}
      </Typography>
      {props.option.status === 'inactive' && (
        <Badge className={styles.inactiveBadge} size='xs'>
          Inactive
        </Badge>
      )}
      {props.option.mask && props.option.mask !== '' && (
        <Badge className={styles.maskBadge} size='xs'>
          {props.option.mask}
        </Badge>
      )}
    </Flex>
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
      {!isMobile && <CommandInput placeholder={props.searchPlaceholder} />}
      <CommandList>
        <CommandEmpty>{props.emptyString}</CommandEmpty>
        {link?.getIsManual() && (
          <CommandGroup heading='Controls'>
            <CommandItem
              onSelect={() => {
                props.setOpen(false);
                showNewBankAccountModal();
              }}
              title='Create an account'
              value='null'
            >
              <div className={styles.controlRow}>
                <CirclePlus className={styles.controlIcon} />
                <Typography className={styles.fullWidth} color='emphasis' ellipsis size='inherit'>
                  Add Another Account
                </Typography>
              </div>
            </CommandItem>
          </CommandGroup>
        )}
        <CommandGroup className={styles.accountsGroup} heading='Accounts'>
          {props.options.map(option => (
            <CommandItem
              key={option.value}
              onSelect={() => {
                props.onSelect?.(option.value);
                props.setOpen(false);
              }}
              title={option.label}
              value={`${option.label} ${option.value}` /* makes search work properly :( */}
            >
              <BankAccountSelectItem currentValue={props.value} option={option} />
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  );
}
