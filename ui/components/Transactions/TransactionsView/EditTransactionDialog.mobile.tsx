import React, { Fragment } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Edit, ExpandMore } from '@mui/icons-material';
import { Accordion, AccordionDetails, AccordionSummary, Button, Dialog, DialogActions, DialogContent, DialogTitle, Divider, FormControl, FormControlLabel, FormLabel, InputBase, List, ListItem, ListItemText, Radio, RadioGroup, Slide, TextField, Typography } from '@mui/material';
import { TransitionProps } from '@mui/material/transitions';
import { Formik, FormikErrors, FormikHelpers } from 'formik';

import TransactionIcon from '../components/TransactionIcon';

import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import useIsMobile from 'hooks/useIsMobile';
import Transaction from 'models/Transaction';
import { useSpending } from 'hooks/spending';


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
  const selectedBankAccountId = useSelectedBankAccountId();
  const spending = useSpending(props.transaction.spendingId);

  const { transaction } = props;

  async function closeDialog() {
    modal.hide()
      .then(() => modal.remove());
  }

  async function validateInput(input: EditTransactionForm): FormikErrors<EditTransactionForm> {
    return {};
  }

  async function submit(values: EditTransactionForm, helper: FormikHelpers<EditTransactionForm>): Promise<void> {
    return Promise.resolve();
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

  function SpentFrom(): JSX.Element {
    let name: JSX.Element | string;
    if (transaction.spendingId && spending) {
      name = <span className="text-semibold">{ spending.name }</span>
    } else if (transaction.spendingId && !spending) {
      name = <span className="text-semibold">...</span>
    } else {
      name = 'Safe-To-Spend';
    }

    return (
      <Fragment>
        <ListItem className='pl-0 pr-0'>
          <Accordion square className='pl-0 pr-0 shadow-none w-full'>
            <AccordionSummary
              style={{
                height: 28 + 16,
                minHeight: 28 + 16,
              }}
              classes={{
                content: 'p-0 m-0 text-xl',
              }}
              className='p-0 m-0'
            >
              <ListItemText
                className='flex-none'
                primary={ 'Spent From' }
                primaryTypographyProps={ {
                  className: 'text-xl',
                } }
              />
              <span className="self-center flex-1 text-end text-xl">
                { name }
              </span>
            </AccordionSummary>
            <AccordionDetails>
              <FormControl>
                <RadioGroup
                  aria-labelledby="demo-radio-buttons-group-label"
                  defaultValue="female"
                  name="radio-buttons-group"
                >
                  <FormControlLabel value="female" control={<Radio />} label="Safe-To-Spend" />
                  <FormControlLabel value="male" control={<Radio />} label="Amazon" />
                  <FormControlLabel value="other" control={<Radio />} label="Vacation" />
                </RadioGroup>
              </FormControl>
            </AccordionDetails>
          </Accordion>
        </ListItem>
        <Divider />
      </Fragment>
    )
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
                  <span className="flex-1 text-end text-xl opacity-75">
                    { transaction.date.format('MMMM Do, YYYY') }
                  </span>
                </EditItem>
                <EditItem name="Status">
                  <span className="flex-1 text-end text-xl opacity-75">
                    { transaction.isPending ? 'Pending' : 'Complete' }
                  </span>
                </EditItem>
                <SpentFrom />
              </List>
            </DialogContent>
            <DialogActions>
              <Button
                color="secondary"
                disabled={ false }
                onClick={ closeDialog }
              >
                Cancel
              </Button>
              <Button
                disabled={ false }
                onClick={ () => {} }
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
