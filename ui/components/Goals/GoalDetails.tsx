import React, { useState } from 'react';
import shallow from 'zustand/shallow';

import CompletedGoalDetails from 'components/Goals/CompletedGoalDetails';
import EditGoalView from 'components/Goals/EditGoalView';
import InProgressGoalDetails from 'components/Goals/InProgressGoalDetails';
import NewGoalDialog from 'components/Goals/NewGoalDialog';
import NoGoals from 'components/Goals/NoGoals';
import TransferDialog from 'components/Spending/TransferDialog';
import { useSelectedGoal } from 'hooks/spending';
import useStore from 'hooks/store';

export default function GoalDetails(): JSX.Element {
  enum DialogOpen {
    NewGoal,
    Transfer,
    EditGoal,
  }
  const [dialog, setDialog] = useState<DialogOpen | null>(null);
  function closeDialog() {
    setDialog(null);
  }

  const goal = useSelectedGoal();

  const { setCurrentGoal } = useStore(state => ({
    setCurrentGoal: state.setCurrentGoal,
  }), shallow);
  function unselectGoal() {
    setCurrentGoal(null);
  }

  function editGoal() {
    setDialog(DialogOpen.EditGoal);
  }

  function openTransferDialog() {
    setDialog(DialogOpen.Transfer);
  }

  function DialogsMaybe(): JSX.Element {
    switch (dialog) {
      case DialogOpen.NewGoal:
        return <NewGoalDialog isOpen onClose={ closeDialog } />;
      case DialogOpen.Transfer:
        return <TransferDialog isOpen onClose={ closeDialog } initialToSpendingId={ goal?.spendingId } />;
      default:
        return null;
    }
  }

  function DetailContents(): JSX.Element {
    if (!goal) {
      return <NoGoals />;
    }

    if (dialog === DialogOpen.EditGoal) {
      return <EditGoalView
        goal={ goal }
        hideView={ closeDialog }
      />;
    }

    if (goal.getGoalIsInProgress()) {
      return <InProgressGoalDetails
        goal={ goal }
        onBack={ unselectGoal }
        openEditView={ editGoal }
        openTransferDialog={ openTransferDialog  }
      />;
    }

    return <CompletedGoalDetails goal={ goal } onBack={ unselectGoal } />;
  }

  return (
    <div className="w-full h-full p-5">
      <DialogsMaybe />
      <DetailContents />
    </div>
  );
}
