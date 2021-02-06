import {LOGOUT} from "../actions";


export default function logout() {
  return dispatch => {
    window.localStorage.removeItem('H-Token');

    return dispatch({
      type: LOGOUT,
    })
  }
}
