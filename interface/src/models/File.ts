import { parseJSON } from 'date-fns';

export default class File {
  fileId: number;
  bankAccountId: number;
  name: string;
  contentType: string;
  size: number;
  createdAt: Date;
  createdByUserId: number;

  constructor(data?: Partial<File>) {
    if (data) {
      Object.assign(this, {
        ...data,
        createdAt: data?.createdAt ?? parseJSON(data?.createdAt),
      });
    }
  }
}
