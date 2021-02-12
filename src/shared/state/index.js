import {combineReducers} from "redux";
import authentication from '../authentication/reducer';
import bootstrap from '../bootstrap/reducer';
import links from '../links/reducer';
import bankAccounts from '../bankAccounts/reducer';

export default combineReducers({
  authentication,
  bankAccounts,
  bootstrap,
  links,
});


