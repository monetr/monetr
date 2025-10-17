import fileSize from '@monetr/interface/util/fileSize';

describe('file size', () => {
  it('will show a human friendly file size', () => {
    {
      // 1KB
      const input = 1024;
      const result = fileSize(input);
      expect(result).toBe('1.0 KiB');
    }

    {
      // 1MB
      const input = 1024 * 1024;
      const result = fileSize(input);
      expect(result).toBe('1.0 MiB');
    }

    {
      // 1GB
      const input = 1024 * 1024 * 1024;
      const result = fileSize(input);
      expect(result).toBe('1.0 GiB');
    }
  });
});
