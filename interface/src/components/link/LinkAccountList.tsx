import { useCallback, useMemo } from 'react';
import { ChevronRight, GripVertical } from 'lucide-react';
import { Link } from 'wouter';

import Badge from '@monetr/interface/components/Badge';
import Typography from '@monetr/interface/components/Typography';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { usePatchLink } from '@monetr/interface/hooks/usePatchLink';
import { type SortableItemProps, useSortableList } from '@monetr/interface/hooks/useSortableList';
import type BankAccount from '@monetr/interface/models/BankAccount';
import type { ID } from '@monetr/interface/models/ID';
import type MonetrLink from '@monetr/interface/models/Link';
import { AmountType } from '@monetr/interface/util/amounts';
import capitalize from '@monetr/interface/util/capitalize';
import sortAccounts from '@monetr/interface/util/sortAccounts';
import { useSnackbar } from '@monetr/notify';

import styles from './LinkAccountList.module.scss';

export interface LinkAccountListProps {
  link: MonetrLink;
  bankAccounts: Array<BankAccount>;
}

export default function LinkAccountList(props: LinkAccountListProps): React.JSX.Element {
  const patchLink = usePatchLink();
  const { enqueueSnackbar } = useSnackbar();

  // The type order from sortAccounts is the default, anything the user hasn't placed yet (like a newly added account)
  // lands here. Archived accounts can't be reordered so they get their own list below.
  const accountsSorted = useMemo(
    () => sortAccounts(props.bankAccounts.filter(bankAccount => !bankAccount.deletedAt)),
    [props.bankAccounts],
  );
  const archived = props.bankAccounts.filter(bankAccount => Boolean(bankAccount.deletedAt));

  const onReorder = useCallback(
    (bankAccountOrder: Array<ID<BankAccount>>) => {
      // The patch updates the link in the cache right away, and puts it back if this fails.
      patchLink({ linkId: props.link.linkId, bankAccountOrder }).catch(
        () =>
          void enqueueSnackbar('Failed to save the order of your accounts.', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      );
    },
    [props.link.linkId, patchLink, enqueueSnackbar],
  );

  const sortable = useSortableList({
    items: accountsSorted,
    order: props.link.bankAccountOrder,
    getId: getBankAccountId,
    onReorder,
  });

  return (
    <section className={styles.root}>
      <div className={styles.header}>
        <Typography color='emphasis' component='h3' size='xl' weight='semibold'>
          Accounts
        </Typography>
        <Badge size='xs'>{accountsSorted.length}</Badge>
      </div>
      <div className={styles.list} ref={sortable.containerRef}>
        {sortable.items.map(bankAccount => (
          <BankAccountItem
            bankAccount={bankAccount}
            key={bankAccount.bankAccountId}
            sortable={sortable.getItemProps(bankAccount.bankAccountId)}
          />
        ))}
      </div>
      {archived.length > 0 && (
        <details className={styles.archived}>
          <summary className={styles.archivedSummary}>Archived ({archived.length})</summary>
          <div className={styles.list}>
            {archived.map(bankAccount => (
              <BankAccountItem bankAccount={bankAccount} key={bankAccount.bankAccountId} />
            ))}
          </div>
        </details>
      )}
    </section>
  );
}

function getBankAccountId(bankAccount: BankAccount): ID<BankAccount> {
  return bankAccount.bankAccountId;
}

interface BankAccountItemProps {
  bankAccount: BankAccount;
  sortable?: SortableItemProps;
}

function BankAccountItem(props: BankAccountItemProps): React.JSX.Element {
  const { data: locale } = useLocaleCurrency(props.bankAccount.currency);
  const path = `/bank/${props.bankAccount.bankAccountId}/settings`;
  return (
    <div {...props.sortable} className={styles.item}>
      <Link className={styles.itemLink} draggable={false} to={path}>
        {Boolean(props.sortable) && <GripVertical aria-hidden='true' className={styles.itemGrip} />}
        <div className={styles.itemText}>
          <div className={styles.itemNameRow}>
            <Typography className={styles.itemName} color='emphasis' ellipsis size='md' weight='semibold'>
              {props.bankAccount.name}
            </Typography>
            {Boolean(props.bankAccount.deletedAt) && <Badge size='sm'>Archived</Badge>}
          </div>
          <Typography color='default' ellipsis size='sm' weight='medium'>
            {capitalize(props.bankAccount.accountSubType)}
          </Typography>
        </div>
        <div className={styles.itemBalance}>
          <Typography color='emphasis' size='md' weight='semibold'>
            {locale?.formatAmount(props.bankAccount.currentBalance, AmountType.Stored)}
          </Typography>
          <Typography color='subtle' size='sm'>
            Available {locale?.formatAmount(props.bankAccount.availableBalance, AmountType.Stored)}
          </Typography>
        </div>
        <ChevronRight className={styles.itemChevron} />
      </Link>
    </div>
  );
}
