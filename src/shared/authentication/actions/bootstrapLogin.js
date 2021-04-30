import User from "data/User";
import request from "shared/util/request";
import { BOOTSTRAP_LOGIN } from "shared/authentication/actions";
import { getAPIUrl } from "shared/bootstrap/selectors";
import axios from "axios";
import Cookies from 'js-cookie'

export default function bootstrapLogin(token = null, user = null) {
  return (dispatch, getState) => {
    if (token) {
      // Trying to switch over to using cookies, but I don't want to break anything at the moment.
      // eslint-disable-next-line no-undef
      if (CONFIG.USE_LOCAL_STORAGE) {
        window.localStorage.setItem('M-Token', token);
      } else {
        Cookies.set('M-Token', token, {
          // eslint-disable-next-line no-undef
          domain: CONFIG.COOKIE_DOMAIN,
          secure: true,
        });
      }
    } else {
      // eslint-disable-next-line no-undef
      if (CONFIG.USE_LOCAL_STORAGE) {
        token = window.localStorage.getItem('H-Token');
      } else {
        token = Cookies.get('M-Token');
      }
    }

    // If the token is not present at this point then the user is not authenticated. We want to dispatch accordingly and
    // store in redux that the user is not authenticated.
    if (!token) {
      dispatch({
        type: BOOTSTRAP_LOGIN,
        payload: {
          isAuthenticated: false,
          token: null,
          user: null,
        }
      });
      return Promise.resolve();
    }

    const apiUrl = getAPIUrl(getState());
    window.API = axios.create({
      baseURL: apiUrl,
      withCredentials: true,
    });

    if (!user) {
      // If we do have the token but we don't have the user info then we need to retrieve it using an API call to get
      // our user data from the API.
      return request().get('/users/me')
        .then(result => {
          dispatch({
            type: BOOTSTRAP_LOGIN,
            payload: {
              isAuthenticated: true,
              token: token,
              user: new User(result.data.user),
            }
          })
        })
        .catch(error => {
          Cookies.remove('M-Token');
          window.localStorage.removeItem('M-Token');
          console.error(error);
        });
    }

    dispatch({
      type: BOOTSTRAP_LOGIN,
      payload: {
        isAuthenticated: true,
        token: token,
        user: new User(user),
      }
    });
    return Promise.resolve();
  };
}
