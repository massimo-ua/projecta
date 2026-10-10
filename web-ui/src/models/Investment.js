export class Investment {
  constructor({
    id,
    key,
    projectId,
    assetId,
    assetName,
    contributor,
    resourceType = 'MONEY',
    amount,
    rawAmount,
    currency,
    homeAmount,
    rawHomeAmount,
    homeCurrency,
    timeHours,
    timeHourlyRate,
    rawTimeHourlyRate,
    goodsQuantity,
    goodsUnit,
    goodsItemName,
    description,
    date,
    rawDate,
    tags = [],
    assetAllocations = [],
  }) {
    this.id = id;
    this.key = key || id;
    this.projectId = projectId;
    this.assetId = assetId;
    this.assetName = assetName;
    this.assetAllocations = Array.isArray(assetAllocations) ? assetAllocations : [];
    this.contributor = contributor;
    this.resourceType = resourceType || 'MONEY';
    this.amount = amount;
    this.rawAmount = rawAmount;
    this.currency = currency;
    this.homeAmount = homeAmount;
    this.rawHomeAmount = rawHomeAmount;
    this.homeCurrency = homeCurrency;
    this.timeHours = timeHours;
    this.timeHourlyRate = timeHourlyRate;
    this.rawTimeHourlyRate = rawTimeHourlyRate;
    this.goodsQuantity = goodsQuantity;
    this.goodsUnit = goodsUnit;
    this.goodsItemName = goodsItemName;
    this.description = description;
    this.date = date;
    this.rawDate = rawDate;
    this.tags = Array.isArray(tags) ? tags : [];
  }

  get isMultiAsset() {
    return this.assetAllocations.length > 1;
  }

  get isMoney() {
    return this.resourceType === 'MONEY';
  }

  get isTime() {
    return this.resourceType === 'TIME';
  }

  get isGoods() {
    return this.resourceType === 'GOODS';
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

  get contributorName() {
    if (!this.contributor) return '';
    return typeof this.contributor === 'string'
      ? this.contributor
      : (this.contributor.displayName || this.contributor.name || '');
  }

  get resourceSummary() {
    if (this.isTime && this.timeHours) {
      const rateStr = this.timeHourlyRate ? ` @ ${this.timeHourlyRate} ${this.currency}/h` : '';
      return `${this.timeHours} hrs${rateStr}`;
    }
    if (this.isGoods && (this.goodsItemName || this.goodsQuantity)) {
      const qtyStr = this.goodsQuantity ? `${this.goodsQuantity} ${this.goodsUnit || 'units'}` : '';
      return [this.goodsItemName, qtyStr].filter(Boolean).join(' • ');
    }
    return 'Capital';
  }
}

export default Investment;
