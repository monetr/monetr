import { Record, Map } from "immutable";


export default class LinksState extends Record({
  items: new Map(),
  loaded: false,
  loading: false,
}) {

}
