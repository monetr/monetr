import { Map } from 'immutable';
import Institution from 'models/Institution';
import Link from 'models/Link';

export default class LinksState {
  constructor() {
    this.items = Map<number, Link>();
  }

  items: Map<number, Link>;
  loaded: boolean;
  loading: boolean;
}

