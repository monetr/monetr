import { type EffectCallback, useEffect } from 'react';

export default function useMountEffect(callback: EffectCallback) {
  useEffect(callback, []);
}
