import { Record, Map } from "immutable";


export default class LinksState extends Record({
  links: new Map(),
  loaded: false,
  loading: false,
}) {

}
