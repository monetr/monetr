import { AxiosInstance } from "axios";

export default function request(): AxiosInstance {
  // @ts-ignore
  // We just want to ignore the type casting.
  return window.API as AxiosInstance;
}
