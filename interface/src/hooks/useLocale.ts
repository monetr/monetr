import { useMemo } from 'react';
import { type UseQueryResult, useQuery } from '@tanstack/react-query';
import type { Locale } from 'date-fns';

import { useCurrentLocale } from '@monetr/interface/hooks/useCurrentLocale';

// localeTable contains a mapping of locale names to an async function to import the date-fns locale data for that
// specific locale. This table is not meant to be used directly and is instead used through the `useLocale` hook. This
// hook will cache the result of the module import such that it only needs to happen once.
const localeTable = {
  af: async () => await import('date-fns/locale/af').then(locale => locale.af),
  ar: async () => await import('date-fns/locale/ar').then(locale => locale.ar),
  arDZ: async () => await import('date-fns/locale/ar-DZ').then(locale => locale.arDZ),
  arEG: async () => await import('date-fns/locale/ar-EG').then(locale => locale.arEG),
  arMA: async () => await import('date-fns/locale/ar-MA').then(locale => locale.arMA),
  arSA: async () => await import('date-fns/locale/ar-SA').then(locale => locale.arSA),
  arTN: async () => await import('date-fns/locale/ar-TN').then(locale => locale.arTN),
  az: async () => await import('date-fns/locale/az').then(locale => locale.az),
  be: async () => await import('date-fns/locale/be').then(locale => locale.be),
  betarask: async () => await import('date-fns/locale/be-tarask').then(locale => locale.beTarask),
  bg: async () => await import('date-fns/locale/bg').then(locale => locale.bg),
  bn: async () => await import('date-fns/locale/bn').then(locale => locale.bn),
  bs: async () => await import('date-fns/locale/bs').then(locale => locale.bs),
  ca: async () => await import('date-fns/locale/ca').then(locale => locale.ca),
  ckb: async () => await import('date-fns/locale/ckb').then(locale => locale.ckb),
  cs: async () => await import('date-fns/locale/cs').then(locale => locale.cs),
  cy: async () => await import('date-fns/locale/cy').then(locale => locale.cy),
  da: async () => await import('date-fns/locale/da').then(locale => locale.da),
  de: async () => await import('date-fns/locale/de').then(locale => locale.de),
  deAT: async () => await import('date-fns/locale/de-AT').then(locale => locale.deAT),
  el: async () => await import('date-fns/locale/el').then(locale => locale.el),
  enAU: async () => await import('date-fns/locale/en-AU').then(locale => locale.enAU),
  enCA: async () => await import('date-fns/locale/en-CA').then(locale => locale.enCA),
  enGB: async () => await import('date-fns/locale/en-GB').then(locale => locale.enGB),
  enIE: async () => await import('date-fns/locale/en-IE').then(locale => locale.enIE),
  enIN: async () => await import('date-fns/locale/en-IN').then(locale => locale.enIN),
  enNZ: async () => await import('date-fns/locale/en-NZ').then(locale => locale.enNZ),
  enUS: async () => await import('date-fns/locale/en-US').then(locale => locale.enUS),
  enZA: async () => await import('date-fns/locale/en-ZA').then(locale => locale.enZA),
  eo: async () => await import('date-fns/locale/eo').then(locale => locale.eo),
  es: async () => await import('date-fns/locale/es').then(locale => locale.es),
  et: async () => await import('date-fns/locale/et').then(locale => locale.et),
  eu: async () => await import('date-fns/locale/eu').then(locale => locale.eu),
  faIR: async () => await import('date-fns/locale/fa-IR').then(locale => locale.faIR),
  fi: async () => await import('date-fns/locale/fi').then(locale => locale.fi),
  fr: async () => await import('date-fns/locale/fr').then(locale => locale.fr),
  frCA: async () => await import('date-fns/locale/fr-CA').then(locale => locale.frCA),
  frCH: async () => await import('date-fns/locale/fr-CH').then(locale => locale.frCH),
  fy: async () => await import('date-fns/locale/fy').then(locale => locale.fy),
  gd: async () => await import('date-fns/locale/gd').then(locale => locale.gd),
  gl: async () => await import('date-fns/locale/gl').then(locale => locale.gl),
  gu: async () => await import('date-fns/locale/gu').then(locale => locale.gu),
  he: async () => await import('date-fns/locale/he').then(locale => locale.he),
  hi: async () => await import('date-fns/locale/hi').then(locale => locale.hi),
  hr: async () => await import('date-fns/locale/hr').then(locale => locale.hr),
  ht: async () => await import('date-fns/locale/ht').then(locale => locale.ht),
  hu: async () => await import('date-fns/locale/hu').then(locale => locale.hu),
  hy: async () => await import('date-fns/locale/hy').then(locale => locale.hy),
  id: async () => await import('date-fns/locale/id').then(locale => locale.id),
  is: async () => await import('date-fns/locale/is').then(locale => locale.is),
  it: async () => await import('date-fns/locale/it').then(locale => locale.it),
  itCH: async () => await import('date-fns/locale/it-CH').then(locale => locale.itCH),
  ja: async () => await import('date-fns/locale/ja').then(locale => locale.ja),
  jaHira: async () => await import('date-fns/locale/ja-Hira').then(locale => locale.jaHira),
  ka: async () => await import('date-fns/locale/ka').then(locale => locale.ka),
  kk: async () => await import('date-fns/locale/kk').then(locale => locale.kk),
  km: async () => await import('date-fns/locale/km').then(locale => locale.km),
  kn: async () => await import('date-fns/locale/kn').then(locale => locale.kn),
  ko: async () => await import('date-fns/locale/ko').then(locale => locale.ko),
  lb: async () => await import('date-fns/locale/lb').then(locale => locale.lb),
  lt: async () => await import('date-fns/locale/lt').then(locale => locale.lt),
  lv: async () => await import('date-fns/locale/lv').then(locale => locale.lv),
  mk: async () => await import('date-fns/locale/mk').then(locale => locale.mk),
  mn: async () => await import('date-fns/locale/mn').then(locale => locale.mn),
  ms: async () => await import('date-fns/locale/ms').then(locale => locale.ms),
  mt: async () => await import('date-fns/locale/mt').then(locale => locale.mt),
  nb: async () => await import('date-fns/locale/nb').then(locale => locale.nb),
  nl: async () => await import('date-fns/locale/nl').then(locale => locale.nl),
  nlBE: async () => await import('date-fns/locale/nl-BE').then(locale => locale.nlBE),
  nn: async () => await import('date-fns/locale/nn').then(locale => locale.nn),
  oc: async () => await import('date-fns/locale/oc').then(locale => locale.oc),
  pl: async () => await import('date-fns/locale/pl').then(locale => locale.pl),
  pt: async () => await import('date-fns/locale/pt').then(locale => locale.pt),
  ptBR: async () => await import('date-fns/locale/pt-BR').then(locale => locale.ptBR),
  ro: async () => await import('date-fns/locale/ro').then(locale => locale.ro),
  ru: async () => await import('date-fns/locale/ru').then(locale => locale.ru),
  se: async () => await import('date-fns/locale/se').then(locale => locale.se),
  sk: async () => await import('date-fns/locale/sk').then(locale => locale.sk),
  sl: async () => await import('date-fns/locale/sl').then(locale => locale.sl),
  sq: async () => await import('date-fns/locale/sq').then(locale => locale.sq),
  sr: async () => await import('date-fns/locale/sr').then(locale => locale.sr),
  srLatn: async () => await import('date-fns/locale/sr-Latn').then(locale => locale.srLatn),
  sv: async () => await import('date-fns/locale/sv').then(locale => locale.sv),
  ta: async () => await import('date-fns/locale/ta').then(locale => locale.ta),
  te: async () => await import('date-fns/locale/te').then(locale => locale.te),
  th: async () => await import('date-fns/locale/th').then(locale => locale.th),
  tr: async () => await import('date-fns/locale/tr').then(locale => locale.tr),
  ug: async () => await import('date-fns/locale/ug').then(locale => locale.ug),
  uk: async () => await import('date-fns/locale/uk').then(locale => locale.uk),
  uz: async () => await import('date-fns/locale/uz').then(locale => locale.uz),
  uzCyrl: async () => await import('date-fns/locale/uz-Cyrl').then(locale => locale.uzCyrl),
  vi: async () => await import('date-fns/locale/vi').then(locale => locale.vi),
  zhCN: async () => await import('date-fns/locale/zh-CN').then(locale => locale.zhCN),
  zhHK: async () => await import('date-fns/locale/zh-HK').then(locale => locale.zhHK),
  zhTW: async () => await import('date-fns/locale/zh-TW').then(locale => locale.zhTW),
};

/**
 * Takes an optional locale name and returns the `date-fns` `Locale` object.
 *
 * @param {string} forceLocale - Force the locale code to be a specific one, otherwise uses user default.
 */
export function useLocale<LocaleName extends keyof typeof localeTable>(
  forceLocale?: LocaleName,
): UseQueryResult<Locale, unknown> {
  const userLocale = useCurrentLocale();
  const requestedLocale = useMemo(() => forceLocale ?? userLocale, [forceLocale, userLocale]);
  return useQuery<Locale, unknown, Locale>({
    queryKey: ['locales', requestedLocale],
    refetchInterval: false,
    refetchOnMount: false,
    refetchOnReconnect: false,
    refetchOnWindowFocus: false,
    async queryFn(context) {
      const [_, localeName] = context.queryKey as Array<string>;
      const [sanitized, language] = sanitizeLocaleName(localeName);
      if (localeName in localeTable) {
        return await localeTable[localeName as LocaleName]();
      }
      if (sanitized in localeTable) {
        return await localeTable[sanitized as LocaleName]();
      }

      if (language in localeTable) {
        return await localeTable[language as LocaleName]();
      }

      throw Error(`Unknown locale name: ${localeName}, not supported within date-fns`);
    },
  });
}

function sanitizeLocaleName(localeName: string): [sanitized: string, languageOnly: string] {
  const sanitized = localeName.replace(/[-|_]/g, '');
  const languageOnly = localeName.slice(0, 2);
  return [sanitized, languageOnly];
}
