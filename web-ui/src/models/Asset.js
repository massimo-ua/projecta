export class Asset {
  constructor({
    id,
    key,
    name,
    description,
    price,
    currency,
    homeAmount,
    homeCurrency,
    type,
    category,
    acquiredAt,
    rawAcquiredAt,
  }) {
    this.id = id;
    this.key = key || id;
    this.name = name;
    this.description = description;
    this.price = price;
    this.currency = currency;
    this.homeAmount = homeAmount;
    this.homeCurrency = homeCurrency;
    this.type = type;
    this.category = category;
    this.acquiredAt = acquiredAt;
    this.rawAcquiredAt = rawAcquiredAt;
  }

  get hasDifferentHomeCurrency() {
    return Boolean(this.homeCurrency && this.currency !== this.homeCurrency);
  }

  get formattedPrice() {
    return `${this.price} ${this.currency}`;
  }

  get formattedHomeAmount() {
    return this.homeAmount && this.hasDifferentHomeCurrency
      ? `≈ ${this.homeAmount} ${this.homeCurrency}`
      : '';
  }
}

export default Asset;
