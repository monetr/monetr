import axios from 'axios';
import request from "shared/util/request";
import { BOOTSTRAP_FAILED, BOOTSTRAP_FINISHED, BOOTSTRAP_START } from "./actions";

function bootstrapStart() {
  return {
    type: BOOTSTRAP_START,
  };
}

function bootstrapFinished(config) {
  return {
    type: BOOTSTRAP_FINISHED,
    config,
  };
}

function bootstrapFailed() {
  return {
    type: BOOTSTRAP_FAILED,
  };
}


export default function bootstrapApplication() {
  return dispatch => {
    dispatch(bootstrapStart());

    // Basically if we need to dynamically get our API URL at runtime instead of at build time we want to send a request
    // to /config.json. This will tell us at runtime where our API is.

    // eslint-disable-next-line no-undef
    if (CONFIG.BOOTSTRAP_CONFIG_JSON) {
      return axios
        .get('/config.json')
        .then(uiConfig => {
          window.API = axios.create({
            baseURL: uiConfig.data.apiUrl,
            withCredentials: true,
          });
          return request().get('/config')
            .then(apiConfig => {
              dispatch(bootstrapFinished({
                ...apiConfig.data,
                ...uiConfig.data,
              }));
            });
        })
        .catch(error => {
          dispatch(bootstrapFailed());
          throw error;
        });
    }

    // But if we already know our API URL from our build configuration then there isn't any work we need to do there and
    // we can just use our build time URL.
    window.API = axios.create({
      // eslint-disable-next-line no-undef
      baseURL: `${CONFIG.API_URL}`,
      withCredentials: true,
    });
    return request().get('/config')
      .then(apiConfig => {
        dispatch(bootstrapFinished({
          ...apiConfig.data,
          ...{
            // eslint-disable-next-line no-undef
            apiUrl: `${CONFIG.API_URL}`,
          },
        }));
      });


  }
}
