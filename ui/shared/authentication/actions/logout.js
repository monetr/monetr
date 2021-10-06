import { LOGOUT } from "shared/authentication/actions";
import Cookies from "js-cookie";

export default function logout() {
  return dispatch => {
    Cookies.remove('M-Token', {
      // eslint-disable-next-line no-undef
      domain: CONFIG.COOKIE_DOMAIN,
      secure: true,
    });
    window.localStorage.removeItem('M-Token');

    return dispatch({
      type: LOGOUT,
    })
  }
}
