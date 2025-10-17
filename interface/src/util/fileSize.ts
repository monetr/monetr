export default function fileSize(bytes: number, si: boolean = false, dp: number = 1) {
  const thresh = si ? 1000 : 1024;

  if (Math.abs(bytes) < thresh) {
    return `${bytes} B`;
  }

  const units = si
    ? ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
    : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
  let x = -1;
  const y = 10 ** dp;

  do {
    bytes /= thresh;
    ++x;
  } while (Math.round(Math.abs(bytes) * y) / y >= thresh && x < units.length - 1);

  return `${bytes.toFixed(dp)} ${units[x]}`;
}
