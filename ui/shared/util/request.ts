import axios from 'axios';
import { AxiosInterface } from 'api/api';

export default function request(): AxiosInterface {
  return axios;
}
