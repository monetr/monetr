import { Store } from "redux";
import { configureStore } from "store";

export function createTestStore(initialState?: any): Store<any, any> {
  return configureStore();
}
