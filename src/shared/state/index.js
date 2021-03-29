import { combineReducers } from "redux";
import authentication from 'shared/authentication/reducer';
import bankAccounts from 'shared/bankAccounts/reducer';
import bootstrap from 'shared/bootstrap/reducer';
import fundingSchedules from 'shared/fundingSchedules/reducer';
import links from 'shared/links/reducer';
import spending from 'shared/spending/reducer';
import transactions from 'shared/transactions/reducer';

export default combineReducers({
  authentication,
  bankAccounts,
  bootstrap,
  fundingSchedules,
  links,
  spending,
  transactions,
});


