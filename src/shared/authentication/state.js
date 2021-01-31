import {Record} from "immutable";
import User from "../../data/user";

export default class AuthenticationState extends Record({
  isAuthenticated: false,
  token: null,
  user: new User(),
}) {

}
