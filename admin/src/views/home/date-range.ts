/** Format range dates in Beijing, with both years shown when crossing a year. */
export function formatOverviewDateRange(start: number, end: number, locale: string, includeYear = false): string {
  const yearFormatter = new Intl.DateTimeFormat(locale, { timeZone: 'Asia/Shanghai', year: 'numeric' });
  const crossesYear = yearFormatter.format(start) !== yearFormatter.format(end);
  const format = (timestamp: number, showYear: boolean) =>
    new Intl.DateTimeFormat(locale, {
      timeZone: 'Asia/Shanghai',
      year: showYear ? 'numeric' : undefined,
      month: locale === 'zh-CN' ? 'long' : 'short',
      day: 'numeric'
    }).format(timestamp);
  return `${format(start, includeYear || crossesYear)} — ${format(end, crossesYear)}`;
}
