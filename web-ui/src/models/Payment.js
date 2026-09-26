export class Payment {
  constructor({
    id,
    key,
    description,
    amount,
    currency,
    homeAmount,
    homeCurrency,
    type,
    category,
    paymentDate,
    kind,
  }) {
    this.id = id;
    this.key = key || id;
    this.description = description;
    this.amount = amount;
    this.currency = currency;
    this.homeAmount = homeAmount;
    this.homeCurrency = homeCurrency;
    this.type = type;
    this.category = category;
    this.paymentDate = paymentDate;
    this.kind = kind;
  }

  get isDownPayment() {
    return this.kind === 'DOWN_PAYMENT';
  }

  get hasDifferentHomeCurrency() {
    return Boolean(this.homeCurrency && this.currency !== this.homeCurrency);
  }

  get formattedAmount() {
    return `${this.amount} ${this.currency}`;
  }

  get formattedHomeAmount() {
    return this.homeAmount && this.hasDifferentHomeCurrency
      ? `≈ ${this.homeAmount} ${this.homeCurrency}`
      : '';
  }
}

export default Payment;
