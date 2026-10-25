package project

import (
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type BudgetFrequency string

const (
	FrequencyOnce     BudgetFrequency = "ONCE"
	FrequencyDaily    BudgetFrequency = "DAILY"
	FrequencyWeekly   BudgetFrequency = "WEEKLY"
	FrequencyMonthly  BudgetFrequency = "MONTHLY"
	FrequencyAnnually BudgetFrequency = "ANNUALLY"
)

func (f BudgetFrequency) String() string { return string(f) }

func ParseBudgetFrequency(val string) (BudgetFrequency, error) {
	switch strings.ToUpper(strings.TrimSpace(val)) {
	case string(FrequencyOnce):
		return FrequencyOnce, nil
	case string(FrequencyDaily):
		return FrequencyDaily, nil
	case string(FrequencyWeekly):
		return FrequencyWeekly, nil
	case string(FrequencyMonthly):
		return FrequencyMonthly, nil
	case string(FrequencyAnnually):
		return FrequencyAnnually, nil
	default:
		return "", exceptions.NewValidationException("invalid budget frequency: must be ONCE, DAILY, WEEKLY, MONTHLY, or ANNUALLY", nil)
	}
}

type BudgetResourceType string

const (
	BudgetResourceMoney BudgetResourceType = "MONEY"
	BudgetResourceTime  BudgetResourceType = "TIME"
	BudgetResourceGoods BudgetResourceType = "GOODS"
)

func (r BudgetResourceType) String() string { return string(r) }

func ParseBudgetResourceType(val string) (BudgetResourceType, error) {
	norm := strings.ToUpper(strings.TrimSpace(val))
	if norm == "" {
		return BudgetResourceMoney, nil
	}
	switch norm {
	case string(BudgetResourceMoney):
		return BudgetResourceMoney, nil
	case string(BudgetResourceTime):
		return BudgetResourceTime, nil
	case string(BudgetResourceGoods):
		return BudgetResourceGoods, nil
	default:
		return "", exceptions.NewValidationException("invalid budget resource type: must be MONEY, TIME, or GOODS", nil)
	}
}

type TagBudgetInput struct {
	Tag            string
	Frequency      string
	TargetMonth    int
	ResourceType   string
	Amount         int64
	TimeHours      *float64
	TimeHourlyRate *int64
	GoodsQuantity  *float64
	GoodsUnit      string
	GoodsItemName  string
	Description    string
}

type TagBudget struct {
	tag            string
	frequency      BudgetFrequency
	targetMonth    int
	resourceType   BudgetResourceType
	amount         int64
	timeHours      *float64
	timeHourlyRate *int64
	goodsQuantity  *float64
	goodsUnit      string
	goodsItemName  string
	description    string
}

func sanitizeTagName(raw string) string {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimLeft(cleaned, "#")
	cleaned = strings.ToLower(strings.TrimSpace(cleaned))
	return cleaned
}

func copyFloatPtr(v *float64) *float64 {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func copyInt64Ptr(v *int64) *int64 {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func NewTagBudget(input TagBudgetInput) (TagBudget, error) {
	tag := sanitizeTagName(input.Tag)
	if tag == "" || utf8.RuneCountInString(tag) > 100 {
		return TagBudget{}, exceptions.NewValidationException("tag name must be between 1 and 100 characters", nil)
	}

	freq, err := ParseBudgetFrequency(input.Frequency)
	if err != nil {
		return TagBudget{}, err
	}

	resType, err := ParseBudgetResourceType(input.ResourceType)
	if err != nil {
		return TagBudget{}, err
	}

	targetMonth := input.TargetMonth
	if targetMonth == 0 {
		targetMonth = 1
	}
	if targetMonth < 1 || targetMonth > 12 {
		return TagBudget{}, exceptions.NewValidationException("target month must be between 1 and 12", nil)
	}

	finalAmount := input.Amount
	if resType == BudgetResourceTime {
		if input.TimeHours == nil || *input.TimeHours <= 0 {
			return TagBudget{}, exceptions.NewValidationException("time hours must be greater than 0 for TIME resource", nil)
		}
		if input.TimeHourlyRate != nil && *input.TimeHourlyRate > 0 && finalAmount <= 0 {
			finalAmount = int64(math.Round(*input.TimeHours * float64(*input.TimeHourlyRate)))
		}
	}
	if resType == BudgetResourceGoods {
		if input.GoodsQuantity != nil && *input.GoodsQuantity <= 0 {
			return TagBudget{}, exceptions.NewValidationException("goods quantity must be greater than 0", nil)
		}
	}
	if finalAmount <= 0 {
		return TagBudget{}, exceptions.NewValidationException("tag budget amount must be greater than 0", nil)
	}

	return TagBudget{
		tag:            tag,
		frequency:      freq,
		targetMonth:    targetMonth,
		resourceType:   resType,
		amount:         finalAmount,
		timeHours:      copyFloatPtr(input.TimeHours),
		timeHourlyRate: copyInt64Ptr(input.TimeHourlyRate),
		goodsQuantity:  copyFloatPtr(input.GoodsQuantity),
		goodsUnit:      strings.TrimSpace(input.GoodsUnit),
		goodsItemName:  strings.TrimSpace(input.GoodsItemName),
		description:    strings.TrimSpace(input.Description),
	}, nil
}

func ReconstituteTagBudget(
	tag string,
	frequency string,
	targetMonth int,
	resourceType string,
	amount int64,
	timeHours *float64,
	timeHourlyRate *int64,
	goodsQuantity *float64,
	goodsUnit string,
	goodsItemName string,
	description string,
) TagBudget {
	freq, err := ParseBudgetFrequency(frequency)
	if err != nil {
		freq = FrequencyMonthly
	}
	resType, err := ParseBudgetResourceType(resourceType)
	if err != nil {
		resType = BudgetResourceMoney
	}
	if targetMonth < 1 || targetMonth > 12 {
		targetMonth = 1
	}
	return TagBudget{
		tag:            sanitizeTagName(tag),
		frequency:      freq,
		targetMonth:    targetMonth,
		resourceType:   resType,
		amount:         amount,
		timeHours:      copyFloatPtr(timeHours),
		timeHourlyRate: copyInt64Ptr(timeHourlyRate),
		goodsQuantity:  copyFloatPtr(goodsQuantity),
		goodsUnit:      goodsUnit,
		goodsItemName:  goodsItemName,
		description:    description,
	}
}

func (b TagBudget) Tag() string                      { return b.tag }
func (b TagBudget) Frequency() BudgetFrequency       { return b.frequency }
func (b TagBudget) TargetMonth() int                 { return b.targetMonth }
func (b TagBudget) ResourceType() BudgetResourceType { return b.resourceType }
func (b TagBudget) Amount() int64                    { return b.amount }
func (b TagBudget) TimeHours() *float64              { return copyFloatPtr(b.timeHours) }
func (b TagBudget) TimeHourlyRate() *int64           { return copyInt64Ptr(b.timeHourlyRate) }
func (b TagBudget) GoodsQuantity() *float64          { return copyFloatPtr(b.goodsQuantity) }
func (b TagBudget) GoodsUnit() string                { return b.goodsUnit }
func (b TagBudget) GoodsItemName() string            { return b.goodsItemName }
func (b TagBudget) Description() string              { return b.description }

// ProjectedAmountForMonth returns the projected minor-unit amount for a given 1-based schedule month index (1..12)
// and the corresponding calendar year/month (used for calendar target month matching and DAILY day count).
func (b TagBudget) ProjectedAmountForMonth(scheduleIndex int, year int, month time.Month) int64 {
	if scheduleIndex < 1 || scheduleIndex > 12 {
		return 0
	}
	switch b.frequency {
	case FrequencyOnce, FrequencyAnnually:
		if int(month) == b.targetMonth {
			return b.amount
		}
		return 0
	case FrequencyMonthly:
		return b.amount
	case FrequencyWeekly:
		return int64(math.Round(float64(b.amount) * 52.0 / 12.0))
	case FrequencyDaily:
		daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
		return b.amount * int64(daysInMonth)
	default:
		return 0
	}
}

// AnnualTotal returns the total projected minor-unit amount across a 12-month schedule starting at startDate up to optional endDate.
func (b TagBudget) AnnualTotal(startDate time.Time, endDate ...time.Time) int64 {
	base := startDate
	if base.IsZero() {
		base = time.Now().UTC()
	}
	var end time.Time
	if len(endDate) > 0 && !endDate[0].IsZero() {
		end = endDate[0]
	}
	startMonth := time.Date(base.Year(), base.Month(), 1, 0, 0, 0, 0, time.UTC)
	var sum int64
	for i := 0; i < 12; i++ {
		m := startMonth.AddDate(0, i, 0)
		if !end.IsZero() && m.After(end) {
			break
		}
		sum += b.ProjectedAmountForMonth(i+1, m.Year(), m.Month())
	}
	return sum
}
