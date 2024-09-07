import { format, formatISO, parseISO } from 'date-fns';
import dayjs from 'dayjs';

export function toDateView(date) {
  return format(parseISO(date), 'dd/MM/yyyy', { awareOfUnicodeTokens: true });
}

export function fromISO(date) {
  return dayjs(date);
}

export function toISO(date) {
  return formatISO(date, { representation: 'complete' });
}

export function toPriceView(price) {
  return fromPrice(price).toFixed(2);
}

export function toPrice(price) {
  return Math.round(price * 100);
}

export function fromPrice(price) {
  return price / 100;
}
