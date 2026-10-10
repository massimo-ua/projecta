export class Asset {
  constructor({
    id,
    key,
    name,
    description,
    status = 'ACTIVE',
    startDate,
    rawStartDate,
    completedDate,
    rawCompletedDate,
    targetPrice,
    rawTargetPrice,
    targetCurrency,
    directCost,
    rawDirectCost,
    totalCost,
    rawTotalCost,
    price,
    rawPrice,
    currency,
    homeAmount,
    homeCurrency,
    type,
    category,
    children = [],
    parents = [],
    acquiredAt,
    rawAcquiredAt,
  }) {
    this.id = id;
    this.key = key || id;
    this.name = name;
    this.description = description;
    this.status = status || 'ACTIVE';
    this.startDate = startDate || acquiredAt;
    this.rawStartDate = rawStartDate || rawAcquiredAt;
    this.completedDate = completedDate;
    this.rawCompletedDate = rawCompletedDate;
    this.targetPrice = targetPrice;
    this.rawTargetPrice = rawTargetPrice;
    this.targetCurrency = targetCurrency || currency;
    this.directCost = directCost;
    this.rawDirectCost = rawDirectCost;
    this.totalCost = totalCost !== undefined ? totalCost : price;
    this.rawTotalCost = rawTotalCost !== undefined ? rawTotalCost : rawPrice;
    this.price = this.totalCost;
    this.rawPrice = this.rawTotalCost;
    this.currency = currency || targetCurrency;
    this.homeAmount = homeAmount;
    this.homeCurrency = homeCurrency;
    this.type = type;
    this.category = category;
    this.children = children || [];
    this.parents = parents || [];
    this.acquiredAt = acquiredAt || startDate;
    this.rawAcquiredAt = rawAcquiredAt || rawStartDate;
  }

  get isActive() {
    return this.status === 'ACTIVE';
  }

  get isCompleted() {
    return this.status === 'COMPLETED';
  }

  get hasTarget() {
    return Boolean(this.rawTargetPrice && this.rawTargetPrice > 0);
  }

  get targetProgress() {
    if (!this.hasTarget || !this.rawTotalCost) return 0;
    const pct = Math.round((this.rawTotalCost / this.rawTargetPrice) * 100);
    return Math.min(Math.max(pct, 0), 100);
  }

  get hasDifferentHomeCurrency() {
    return Boolean(this.homeCurrency && this.currency !== this.homeCurrency);
  }

  get formattedPrice() {
    return `${this.price || this.totalCost || '0'} ${this.currency || ''}`.trim();
  }

  get formattedDirectCost() {
    return `${this.directCost || '0'} ${this.currency || ''}`.trim();
  }

  get formattedTotalCost() {
    return `${this.totalCost || '0'} ${this.currency || ''}`.trim();
  }

  get formattedTargetPrice() {
    return this.hasTarget ? `${this.targetPrice} ${this.targetCurrency}` : '';
  }

  get formattedHomeAmount() {
    return this.homeAmount && this.hasDifferentHomeCurrency
      ? `≈ ${this.homeAmount} ${this.homeCurrency}`
      : '';
  }

  get hasChildren() {
    return Array.isArray(this.children) && this.children.length > 0;
  }

  get hasParents() {
    return Array.isArray(this.parents) && this.parents.length > 0;
  }
}

export default Asset;
