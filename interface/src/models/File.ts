import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class File {
  fileId: string;
  name: string;
  contentType: string;
  size: number;
  createdAt: Date;
  createdByUserId: number;

  constructor(data: WithJsonValues<File>) {
    this.fileId = data.fileId;
    this.name = data.name;
    this.contentType = data.contentType;
    this.size = data.size;
    this.createdAt = parseDate(data.createdAt);
    this.createdByUserId = data.createdByUserId;
  }
}
