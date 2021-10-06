import {Record} from "immutable";
import Login from "data/Login";

export default class User extends Record({
  userId: 0,
  loginId: 0,
  login: new Login(),
  accountId: 0,
  firstName: '',
  lastName: '',
}) {

}
