import {combineReducers} from "redux";
import authentication from '../authentication/reducer';
import bootstrap from '../bootstrap/reducer';

export default combineReducers({
  authentication,
  bootstrap,
});


