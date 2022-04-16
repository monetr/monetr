export function bigIntToHex(input: bigint): string {
  return input.toString(16);
}

export function hexToBigInt(hex: string): bigint {
  return BigInt(`0x${ hex.replace('0x', '') }`)
}

export function hexToUint8Array(hexString: string): Uint8Array {
  if (hexString === undefined) {
    throw RangeError('hexString cannot undefined')
  }

  const hexMatch = hexString.match(/^(0x)?([\da-fA-F]+)$/)
  if (hexMatch == null) {
    throw RangeError('hexString must be a hexadecimal string, e.g. \'0x4dc43467fe91\' or \'4dc43467fe91\'')
  }

  let hex = hexMatch[2]
  hex = (hex.length % 2 === 0) ? hex : '0' + hex

  return Uint8Array.from(hex.match(/[\da-fA-F]{2}/g)!.map((h) => parseInt(h, 16)));
}

export function uint8ArrayToHex(array: Uint8Array): string {
  return Array.from(array, byte => ('0' + (byte & 0xFF).toString(16)).slice(-2)).join('');
}
