import {combineReducers} from "redux";
import authentication from '../authentication/reducer';
import bootstrap from '../bootstrap/reducer';
import links from '../links/reducer';

export default combineReducers({
  authentication,
  bootstrap,
  links,
});


