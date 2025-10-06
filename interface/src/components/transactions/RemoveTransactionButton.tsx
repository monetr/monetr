import React, { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Trash } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import { showRemoveTransactionModal } from '@monetr/interface/modals/RemoveTransactionModal';
import Transaction from '@monetr/interface/models/Transaction';

interface RemoveTransactionButtonProps {
  transaction: Transaction;
}

export default function RemoveTransactionButton(props: RemoveTransactionButtonProps): JSX.Element {
  const { transaction } = props;
  const { data: link } = useCurrentLink();
  const navigate = useNavigate();

  const promptRemoveTransaction = useCallback(async () => {
    return await showRemoveTransactionModal({
      transaction,
    })
      .then(() => navigate(`/bank/${ transaction.bankAccountId }/transactions`));
  }, [navigate, transaction]);

  // We only allow removing transactions on manual links.
  if (!link.getIsManual()) {
    return null;
  }

  return (
    <Button
      variant='destructive'
      onClick={ promptRemoveTransaction }
    >
      <Trash />
      Remove
    </Button>
  );
}
