import React, { Fragment, useState } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Edit, ExpandMore, Inbox, Inbox, Mail, Mail } from '@mui/icons-material';
import { Accordion, AccordionDetails, AccordionSummary, Box, Button, ButtonBase, Dialog, DialogActions, DialogContent, DialogTitle, Divider, FormControl, FormControlLabel, FormLabel, InputBase, List, ListItem, ListItemButton, ListItemIcon, ListItemText, Radio, RadioGroup, Slide, SwipeableDrawer, TextField, Typography } from '@mui/material';
import { TransitionProps } from '@mui/material/transitions';
import { Formik, FormikErrors, FormikHelpers } from 'formik';

import TransactionIcon from '../components/TransactionIcon';

import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import useIsMobile from 'hooks/useIsMobile';
import Transaction from 'models/Transaction';
import { useSpending, useSpendingSink } from 'hooks/spending';
import TransactionSpentFromSelectionMobile from './TransactionSpentFromSelection.mobile';
import clsx from 'clsx';
import { AxiosError } from 'axios';
import { useUpdateTransaction } from 'hooks/transactions';


export interface EditTransactionDialogMobileProps {
  transaction: Transaction;
}

interface EditTransactionForm {
  name: string;
  spendingId: number | null;
}

const Transition = React.forwardRef(function Transition(
  props: TransitionProps & {
    children: React.ReactElement<any, any>;
  },
  ref: React.Ref<unknown>,
) {
  return <Slide direction="left" ref={ ref } { ...props } />;
});

function EditTransactionDialogMobile(props: EditTransactionDialogMobileProps): JSX.Element {
  const modal = useModal();
  const isMobile = useIsMobile();
  const updateTransaction = useUpdateTransaction();
  const { result: allSpending } = useSpendingSink();
  const [drawerOpen, setDrawerOpen] = useState<boolean>(false);

  const { transaction } = props;

  async function closeDialog() {
    modal.hide()
      .then(() => modal.remove());
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
        alert(error?.response?.data?.error || 'Could not save transaction name.')
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
          { children }
        </ListItem>
        <Divider />
      </Fragment>
    );
  }

  function spendingName(spendingId: number | null): JSX.Element | string {
    if (spendingId === null) {
      return 'Safe-To-Spend';
    }
    const spending = allSpending.find(item => item.spendingId === spendingId);
    if (spending) {
      return <span className="text-semibold">{ spending.name }</span>
    }

    return <span className="text-semibold">...</span>
  }

  return (
    <Dialog
      open={ modal.visible }
      maxWidth="sm"
      fullScreen={ isMobile }
      TransitionComponent={ Transition }
      keepMounted={ false }
    >
      <Formik
        initialValues={ initialValues }
        validate={ validateInput }
        onSubmit={ submit }
      >
        { ({
          values,
          errors,
          touched,
          handleChange,
          handleBlur,
          handleSubmit,
          setFieldValue,
          isSubmitting,
          submitForm,
          isValid,
        }) => (
          <Fragment>
            <DialogTitle>
              <div className='w-full flex justify-center'>
                <TransactionIcon transaction={ transaction } size={ 80 }  />
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
                  <span className="flex-1 text-end text-xl opacity-75">
                    { transaction.getOriginalName() }
                  </span>
                </EditItem>
                <EditItem name="Date">
                  <span className={ clsx("flex-1 text-end text-xl opacity-75",{
                    'text-green-600': props.transaction.getIsAddition(),
                    'text-red-600': !props.transaction.getIsAddition(),
                  })}>
                    { transaction.getAmountString() }
                  </span>
                </EditItem>
                <EditItem name="Date">
                  <span className="flex-1 text-end text-xl opacity-75">
                    { transaction.date.format('MMMM Do, YYYY') }
                  </span>
                </EditItem>
                <EditItem name="Status">
                  <span className="flex-1 text-end text-xl opacity-75">
                    { transaction.isPending ? 'Pending' : 'Complete' }
                  </span>
                </EditItem>
                { !transaction.getIsAddition() && 
                  <Fragment>
                    <TransactionSpentFromSelectionMobile
                      open={ drawerOpen }
                      onClose={ () => setDrawerOpen(false) }
                      onChange={ (value) => setFieldValue('spendingId', value) }
                      value={ values.spendingId }
                    />
                    <ListItemButton
                      className='pl-0 pr-0'
                      onClick={ () => setDrawerOpen(true) }
                      disabled={ isSubmitting }
                    >
                      <ListItemText
                        className='flex-none'
                        primary={ "Spent From" }
                        primaryTypographyProps={ {
                          className: 'text-xl',
                        } }
                      />
                      <span className="flex-1 text-end text-xl">
                        { spendingName(values.spendingId) }
                      </span>
                    </ListItemButton>
                    <Divider />
                  </Fragment>
                }
              </List>
            </DialogContent>
            <DialogActions>
              <Button
                color="secondary"
                disabled={ isSubmitting }
                onClick={ closeDialog }
              >
                Cancel
              </Button>
              <Button
                disabled={ isSubmitting }
                onClick={ submitForm }
                color="primary"
                type="submit"
              >
                Save
              </Button>
            </DialogActions>
          </Fragment>
        ) }
      </Formik>
    </Dialog>
  );
}

const editTransactionMobileModal = NiceModal.create(EditTransactionDialogMobile);
export default editTransactionMobileModal;

export function showEditTransactionMobileDialog(props: EditTransactionDialogMobileProps): void {
  NiceModal.show(editTransactionMobileModal, props);
}
