import { AppStore, configureStore } from 'store';

export function createTestStore(): AppStore {
  return configureStore();
}
