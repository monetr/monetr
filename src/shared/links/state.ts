import { Map } from "immutable";
import Link from "data/Link";


export default class LinksState {
  constructor() {
    this.items = Map<number, Link>();
  }

  items: Map<number, Link>;
  loaded: boolean;
  loading: boolean;
}

