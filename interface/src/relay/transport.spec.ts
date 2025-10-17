import { makeSneakyFetchTransport } from '@monetr/interface/relay/transport';

describe('sentry relay transport', () => {
  it('will make a sneaky transport', () => {
    const { send, flush } = makeSneakyFetchTransport(
      {
        fetchOptions: undefined,
        url: 'http://my.monetr.dev',
        recordDroppedEvent: jest.fn(),
        headers: {
          Key: 'value',
        },
      },
      jest.fn(),
    );

    // TODO This is fine for now but eventually I'd love to just mock the actual make transport function inside so we
    // can make sure that its being called with the correct options.
    expect(typeof send).toBe('function');
    expect(typeof flush).toBe('function');
  });
});
