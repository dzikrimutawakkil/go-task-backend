package workspaces

import "gorm.io/gorm"

// TierPlanRepository handles database operations for tier_plans and tier_limits tables.
// M5: Subscription Tiers — Phase 1: Database & Models
type TierPlanRepository interface {
	FindAllActivePlans() ([]TierPlan, error)
	FindPlanByTier(tier string) (*TierPlan, error)
	FindLimitsByTier(tier string) (*TierLimit, error)
	FindAllPlansWithLimits() ([]TierPlanWithLimits, error)
}

type tierPlanRepository struct {
	db *gorm.DB
}

func NewTierPlanRepository(db *gorm.DB) TierPlanRepository {
	return &tierPlanRepository{db}
}

// FindAllActivePlans returns all active tier plans.
func (r *tierPlanRepository) FindAllActivePlans() ([]TierPlan, error) {
	var plans []TierPlan
	err := r.db.Where("is_active = ?", true).Find(&plans).Error
	return plans, err
}

// FindPlanByTier returns a tier plan by its tier name.
func (r *tierPlanRepository) FindPlanByTier(tier string) (*TierPlan, error) {
	var plan TierPlan
	err := r.db.Where("tier = ?", tier).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// FindLimitsByTier returns tier limits for a specific tier.
func (r *tierPlanRepository) FindLimitsByTier(tier string) (*TierLimit, error) {
	var limits TierLimit
	err := r.db.Where("tier = ?", tier).First(&limits).Error
	if err != nil {
		return nil, err
	}
	return &limits, nil
}

// FindAllPlansWithLimits returns all active plans with their limits.
func (r *tierPlanRepository) FindAllPlansWithLimits() ([]TierPlanWithLimits, error) {
	var plans []TierPlan
	if err := r.db.Where("is_active = ?", true).Find(&plans).Error; err != nil {
		return nil, err
	}

	var result []TierPlanWithLimits
	for _, plan := range plans {
		var limits TierLimit
		if err := r.db.Where("tier = ?", plan.Tier).First(&limits).Error; err != nil {
			continue
		}

		result = append(result, TierPlanWithLimits{
			Tier:         plan.Tier,
			Name:         plan.Name,
			Description:  plan.Description,
			PriceMonthly: plan.PriceMonthly,
			PriceYearly:  plan.PriceYearly,
			Limits:       limits,
			Features: Features{
				Comments: limits.CanComment,
				Realtime: limits.CanSSE,
				AuditLog: limits.CanAuditLog,
			},
		})
	}

	return result, nil
}
