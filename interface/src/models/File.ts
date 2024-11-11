import parseDate from '@monetr/interface/util/parseDate';

export default class File {
  fileId: string;
  name: string;
  contentType: string;
  size: number;
  createdAt: Date;
  createdByUserId: number;

  constructor(data?: Partial<File>) {
    if (data) {
      Object.assign(this, {
        ...data,
        createdAt: parseDate(data?.createdAt),
      });
    }
  }
}
