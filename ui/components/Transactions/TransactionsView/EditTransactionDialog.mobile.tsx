import React, { Fragment, useState } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Button, DialogActions, DialogContent, DialogTitle, Divider, InputBase, List, ListItem, ListItemButton, ListItemText, SwipeableDrawer } from '@mui/material';
import { AxiosError } from 'axios';
import { Formik, FormikErrors, FormikHelpers } from 'formik';

import TransactionSpentFromSelectionMobile from './TransactionSpentFromSelection.mobile';

import TransactionIcon from '../components/TransactionIcon';

import clsx from 'clsx';
import VerticalPuller from 'components/VerticalPuller';
import { useSpendingSink } from 'hooks/spending';
import { useUpdateTransaction } from 'hooks/transactions';
import Transaction from 'models/Transaction';


export interface EditTransactionDialogMobileProps {
  transaction: Transaction;
}

interface EditTransactionForm {
  name: string;
  spendingId: number | null;
}

function EditTransactionDialogMobile(props: EditTransactionDialogMobileProps): JSX.Element {
  const modal = useModal();
  const updateTransaction = useUpdateTransaction();
  const { result: allSpending } = useSpendingSink();
  const [drawerOpen, setDrawerOpen] = useState<boolean>(false);

  const { transaction } = props;

  async function closeDialog() {
    modal.keepMounted = false;
    modal.hide();
    setTimeout(() => modal.remove(), 500);
    // TODO The promise from `modal.hide()` never resolves, so modal.remove is never called; causing issues with
    //   the same modal being re-used here.
    // modal.hide()
    //   .then(() => console.warn(modal.remove()));
  }

  async function validateInput(input: EditTransactionForm): FormikErrors<EditTransactionForm> {
    return {};
  }

  async function submit(values: EditTransactionForm, helper: FormikHelpers<EditTransactionForm>): Promise<void> {
    helper.setSubmitting(true);
    const updatedTransaction = new Transaction({
      ...transaction,
      spendingId: values.spendingId,
      name: values.name,
    });

    return updateTransaction(updatedTransaction)
      .then(() => closeDialog())
      .catch((error: AxiosError) => {
        alert(error?.response?.data?.error || 'Could not save transaction name.');
        helper.setSubmitting(false);
      });
  }

  const initialValues: EditTransactionForm = {
    name: transaction.getName(),
    spendingId: transaction.spendingId,
  };

  function onFocus(e: React.FocusEvent<any>) {
    setTimeout(() => {
      e.target.setSelectionRange(e.target.value.length, e.target.value.length);
    });
  }

  function EditItem({ name, children }): JSX.Element {
    return (
      <Fragment>
        <ListItem className='pl-0 pr-0'>
          <ListItemText
            className='flex-none'
            primary={ name }
            primaryTypographyProps={ {
              className: 'text-xl',
            } }
          />
          {children}
        </ListItem>
        <Divider />
      </Fragment>
    );
  }

  function spendingName(spendingId: number | null): JSX.Element | string {
    if (spendingId === null) {
      return 'Free-To-Use';
    }
    const spending = allSpending.find(item => item.spendingId === spendingId);
    if (spending) {
      return <span className="text-semibold">{spending.name}</span>;
    }

    return <span className="text-semibold">...</span>;
  }

  return (
    <SwipeableDrawer
      transitionDuration={ 250 }
      style={ {
        zIndex: 1100,
      } }
      anchor='right'
      open={ modal.visible }
      onClose={ closeDialog }
      onOpen={ () => { } }
    >
      <div className='h-full flex flex-col w-[100vw]'>
        <VerticalPuller />
        <Formik
          initialValues={ initialValues }
          validate={ validateInput }
          onSubmit={ submit }
        >
          {({
            values,
            handleChange,
            handleBlur,
            setFieldValue,
            isSubmitting,
            submitForm,
          }) => (
            <Fragment>
              <DialogTitle>
                <div className='w-full flex justify-center'>
                  <TransactionIcon transaction={ transaction } size={ 80 } />
                </div>
              </DialogTitle>
              <DialogContent>
                <List className='pl-0 pr-0'>
                  <Divider />
                  <EditItem name="Name">
                    <InputBase
                      name='name'
                      className='flex-1 flex text-end'
                      disabled={ isSubmitting }
                      style={ { height: 28 } }
                      onChange={ handleChange }
                      onBlur={ handleBlur }
                      onFocus={ onFocus }
                      value={ values.name }
                      inputProps={ {
                        className: 'flex text-end text-xl',
                      } }
                    />
                  </EditItem>
                  <EditItem name="Original Name">
                    <span className="flex-1 text-end text-xl opacity-60">
                      {transaction.getOriginalName()}
                    </span>
                  </EditItem>
                  <EditItem name="Date">
                    <span className={ clsx('flex-1 text-end text-xl opacity-60', {
                      'text-green-600': props.transaction.getIsAddition(),
                      'text-red-600': !props.transaction.getIsAddition(),
                    }) }>
                      {transaction.getAmountString()}
                    </span>
                  </EditItem>
                  <EditItem name="Date">
                    <span className="flex-1 text-end text-xl opacity-60">
                      {transaction.date.format('MMMM Do, YYYY')}
                    </span>
                  </EditItem>
                  <EditItem name="Status">
                    <span className="flex-1 text-end text-xl opacity-60">
                      {transaction.isPending ? 'Pending' : 'Complete'}
                    </span>
                  </EditItem>
                  {!transaction.getIsAddition() &&
                    <Fragment>
                      <TransactionSpentFromSelectionMobile
                        open={ drawerOpen }
                        onClose={ () => setDrawerOpen(false) }
                        onChange={ value => setFieldValue('spendingId', value) }
                        value={ values.spendingId }
                      />
                      <ListItemButton
                        className='pl-0 pr-0'
                        onClick={ () => setDrawerOpen(true) }
                        disabled={ isSubmitting }
                      >
                        <ListItemText
                          className='flex-none'
                          primary="Spent From"
                          primaryTypographyProps={ {
                            className: 'text-xl',
                          } }
                        />
                        <span className="flex-1 text-end text-xl">
                          {spendingName(values.spendingId)}
                        </span>
                      </ListItemButton>
                      <Divider />
                    </Fragment>
                  }
                </List>
              </DialogContent>
              <DialogActions className="bg-purple-900">
                <Button
                  color="secondary"
                  disabled={ isSubmitting }
                  onClick={ closeDialog }
                  variant="outlined"
                >
                  Cancel
                </Button>
                <Button
                  disabled={ isSubmitting }
                  onClick={ submitForm }
                  color="primary"
                  type="submit"
                  variant="contained"
                >
                  Save
                </Button>
              </DialogActions>
            </Fragment>
          )}
        </Formik>
      </div>
    </SwipeableDrawer>
  );
}

const editTransactionMobileModal = NiceModal.create(EditTransactionDialogMobile);
export default editTransactionMobileModal;

export function showEditTransactionMobileDialog(props: EditTransactionDialogMobileProps): void {
  NiceModal.show(editTransactionMobileModal, props);
}
