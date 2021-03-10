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
    return axios
      .get('/config.json')
      .then(uiConfig => {
        window.API = axios.create({
          baseURL: uiConfig.data.apiUrl,
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
}
