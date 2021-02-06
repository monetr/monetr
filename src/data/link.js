import {Record} from "immutable";

export default class Link extends Record({
  linkId: 0,
  institutionId: '',
  institutionName: '',
  customInstitutionName: null,
  createdByUserId: 0
}) {
  linkId;
  institutionId;
  institutionName;
  customInstitutionName;
  createdByUserId;

  getName() {
    return this.customInstitutionName || this.institutionName;
  }
}

