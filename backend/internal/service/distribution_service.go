package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrDistributionCodeInvalid      = infraerrors.BadRequest("DISTRIBUTION_CODE_INVALID", "invalid agent promotion code")
	ErrDistributionCodeConflict     = infraerrors.BadRequest("PROMOTION_BINDING_CONFLICT", "invitation rebate code and agent promotion code cannot be used together")
	ErrDistributionAlreadyBound     = infraerrors.Conflict("DISTRIBUTION_ALREADY_BOUND", "distribution agent already bound")
	ErrDistributionNotAgent         = infraerrors.Forbidden("DISTRIBUTION_NOT_AGENT", "user is not an agent")
	ErrDistributionDisabled         = infraerrors.Forbidden("DISTRIBUTION_DISABLED", "distribution is disabled")
	ErrSubagentRecruitmentDenied    = infraerrors.Forbidden("SUBAGENT_RECRUITMENT_DENIED", "subagent recruitment permission is required")
	ErrSubagentCandidateUnavailable = infraerrors.BadRequest("SUBAGENT_CANDIDATE_UNAVAILABLE", "unable to add this user; verify the email and user eligibility")
	ErrPromotionTrackingInput       = infraerrors.BadRequest("INVALID_PROMOTION_TRACKING_INPUT", "invalid promotion tracking input")
	ErrPromotionArchiveUnavailable  = infraerrors.BadRequest("PROMOTION_ARCHIVE_UNAVAILABLE", "the requested archived promotion range cannot be reported exactly")
)

type DistributionAgent struct {
	ID                     int64           `json:"id"`
	UserID                 int64           `json:"user_id"`
	LevelID                int64           `json:"level_id"`
	Depth                  int             `json:"depth"`
	ParentAgentID          *int64          `json:"parent_agent_id,omitempty"`
	PromotionCode          string          `json:"promotion_code"`
	RateOverrideBPS        *int            `json:"rate_override_bps,omitempty"`
	EffectiveRateBPS       int             `json:"effective_rate_bps"`
	MaxChildRateBPS        int             `json:"max_child_rate_bps"`
	CanRecruitSubagents    bool            `json:"can_recruit_subagents"`
	CanViewPromotionStats  bool            `json:"can_view_promotion_stats"`
	Status                 string          `json:"status"`
	AvailableCNY           decimal.Decimal `json:"available_cny"`
	FrozenCNY              decimal.Decimal `json:"frozen_cny"`
	ReservedCNY            decimal.Decimal `json:"reserved_cny"`
	DebtCNY                decimal.Decimal `json:"debt_cny"`
	TotalEarnedCNY         decimal.Decimal `json:"total_earned_cny"`
	TotalWithdrawnCNY      decimal.Decimal `json:"total_withdrawn_cny"`
	Email                  string          `json:"email,omitempty"`
	Username               string          `json:"username,omitempty"`
	UserStatus             string          `json:"user_status,omitempty"`
	ParentUserID           int64           `json:"parent_user_id,omitempty"`
	ParentEmail            string          `json:"parent_email,omitempty"`
	ParentUsername         string          `json:"parent_username,omitempty"`
	CustomerCount          int64           `json:"customer_count,omitempty"`
	TeamCount              int64           `json:"team_count,omitempty"`
	PayingCustomerCount    int64           `json:"paying_customer_count"`
	CustomerPaidCNY        decimal.Decimal `json:"customer_paid_cny"`
	ThisMonthCommissionCNY decimal.Decimal `json:"this_month_commission_cny"`
	TotalConvertedCNY      decimal.Decimal `json:"total_converted_cny"`
	LastCommissionAt       *time.Time      `json:"last_commission_at,omitempty"`
	CreatedAt              time.Time       `json:"created_at,omitempty"`
}

type DistributionAdminListFilter struct {
	Page      int
	PageSize  int
	Search    string
	Status    string
	Depth     int
	AgentID   int64
	SortBy    string
	SortOrder string
}

type DistributionAdminCommissionListFilter struct {
	Page        int
	PageSize    int
	Search      string
	Status      string
	EntryType   string
	PaymentType string
	AgentID     int64
	DateFrom    *time.Time
	DateTo      *time.Time
	SortBy      string
	SortOrder   string
}

type DistributionAdminWithdrawalListFilter struct {
	Page      int
	PageSize  int
	Search    string
	Status    string
	AgentID   int64
	DateFrom  *time.Time
	DateTo    *time.Time
	SortBy    string
	SortOrder string
}

type DistributionAdminAnomalyListFilter struct {
	Page      int
	PageSize  int
	Type      string
	Severity  string
	Search    string
	SortOrder string
}

type DistributionUserListFilter struct {
	Page      int
	PageSize  int
	Search    string
	Status    string
	EntryType string
}

type DistributionUserOption struct {
	UserID            int64  `json:"user_id"`
	Email             string `json:"email"`
	Username          string `json:"username"`
	Status            string `json:"status"`
	Selectable        bool   `json:"selectable"`
	UnavailableReason string `json:"unavailable_reason,omitempty"`
}

type DistributionAgentOption struct {
	AgentID       int64  `json:"agent_id"`
	UserID        int64  `json:"user_id"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	PromotionCode string `json:"promotion_code"`
	Depth         int    `json:"depth"`
	Status        string `json:"status"`
}

type DistributionAgentEvent struct {
	ID                  int64     `json:"id"`
	AgentID             int64     `json:"agent_id"`
	EventType           string    `json:"event_type"`
	OldStatus           string    `json:"old_status,omitempty"`
	NewStatus           string    `json:"new_status,omitempty"`
	OldRateOverrideBPS  *int      `json:"old_rate_override_bps,omitempty"`
	NewRateOverrideBPS  *int      `json:"new_rate_override_bps,omitempty"`
	OldEffectiveRateBPS *int      `json:"old_effective_rate_bps,omitempty"`
	NewEffectiveRateBPS *int      `json:"new_effective_rate_bps,omitempty"`
	Reason              string    `json:"reason"`
	ActorUserID         *int64    `json:"actor_user_id,omitempty"`
	ActorEmail          string    `json:"actor_email,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

type DistributionBindingEvent struct {
	ID               int64     `json:"id"`
	CustomerUserID   int64     `json:"customer_user_id"`
	OldAgentID       *int64    `json:"old_agent_id,omitempty"`
	NewAgentID       int64     `json:"new_agent_id"`
	OldAgentEmail    string    `json:"old_agent_email,omitempty"`
	NewAgentEmail    string    `json:"new_agent_email"`
	OldPromotionCode string    `json:"old_promotion_code,omitempty"`
	NewPromotionCode string    `json:"new_promotion_code"`
	Reason           string    `json:"reason"`
	ActorUserID      *int64    `json:"actor_user_id,omitempty"`
	ActorEmail       string    `json:"actor_email,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type DistributionCommissionInput struct {
	PaymentOrderID   int64
	CustomerUserID   int64
	PaymentType      string
	PaymentCurrency  string
	ActualPaid       decimal.Decimal
	ProviderSnapshot json.RawMessage
	PaidAt           time.Time
}

type DistributionCommissionResult struct {
	SourceID   int64           `json:"source_id"`
	Status     string          `json:"status"`
	BaseCNY    decimal.Decimal `json:"base_cny"`
	Commission decimal.Decimal `json:"commission_cny"`
	Created    bool            `json:"created"`
}

type DistributionOverview struct {
	Agent                    *DistributionAgent `json:"agent"`
	DistributionEnabled      bool               `json:"distribution_enabled"`
	PromotionTrackingEnabled bool               `json:"promotion_tracking_enabled"`
	CustomerCount            int64              `json:"customer_count"`
	TeamCount                int64              `json:"team_count"`
	PayingCustomerCount      int64              `json:"paying_customer_count"`
	NewCustomersThisMonth    int64              `json:"new_customers_this_month"`
	CustomerPaidCNY          decimal.Decimal    `json:"customer_paid_cny"`
	ThisMonthCustomerPaid    decimal.Decimal    `json:"this_month_customer_paid_cny"`
	ThisMonthCommissionCNY   decimal.Decimal    `json:"this_month_commission_cny"`
}

type DistributionAdminOverview struct {
	TotalAgents            int64           `json:"total_agents"`
	ActiveAgents           int64           `json:"active_agents"`
	SuspendedAgents        int64           `json:"suspended_agents"`
	NewAgentsThisMonth     int64           `json:"new_agents_this_month"`
	TotalCustomers         int64           `json:"total_customers"`
	PayingCustomers        int64           `json:"paying_customers"`
	NewCustomersThisMonth  int64           `json:"new_customers_this_month"`
	CustomerPaidCNY        decimal.Decimal `json:"customer_paid_cny"`
	TotalCommissionCNY     decimal.Decimal `json:"total_commission_cny"`
	MonthCommissionCNY     decimal.Decimal `json:"month_commission_cny"`
	ReversedCommissionCNY  decimal.Decimal `json:"reversed_commission_cny"`
	AvailableCommissionCNY decimal.Decimal `json:"available_commission_cny"`
	FrozenCommissionCNY    decimal.Decimal `json:"frozen_commission_cny"`
	ReservedCommissionCNY  decimal.Decimal `json:"reserved_commission_cny"`
	DebtCommissionCNY      decimal.Decimal `json:"debt_commission_cny"`
	PendingWithdrawals     int64           `json:"pending_withdrawals"`
	PendingWithdrawalCNY   decimal.Decimal `json:"pending_withdrawal_cny"`
	PayingWithdrawals      int64           `json:"paying_withdrawals"`
	PayingWithdrawalCNY    decimal.Decimal `json:"paying_withdrawal_cny"`
	PaidThisMonthCNY       decimal.Decimal `json:"paid_this_month_cny"`
	OverduePendingCount    int64           `json:"overdue_pending_count"`
}

type DistributionSettings struct {
	Enabled                      bool            `json:"enabled"`
	L1DefaultRateBPS             int             `json:"l1_default_rate_bps"`
	L1MaxChildRateBPS            int             `json:"l1_max_child_rate_bps"`
	L2DefaultRateBPS             int             `json:"l2_default_rate_bps"`
	FreezeHours                  int             `json:"freeze_hours"`
	WithdrawalEnabled            bool            `json:"withdrawal_enabled"`
	WithdrawalDualApproval       bool            `json:"withdrawal_dual_approval_enabled"`
	MinimumWithdrawalCNY         decimal.Decimal `json:"minimum_withdrawal_cny"`
	MaximumWithdrawalCNY         decimal.Decimal `json:"maximum_withdrawal_cny"`
	WithdrawalFeeRateBPS         int             `json:"withdrawal_fee_rate_bps"`
	WithdrawalFeeFixedCNY        decimal.Decimal `json:"withdrawal_fee_fixed_cny"`
	DailyWithdrawalLimitCNY      decimal.Decimal `json:"daily_withdrawal_limit_cny"`
	MonthlyWithdrawalLimitCNY    decimal.Decimal `json:"monthly_withdrawal_limit_cny"`
	CNYPerPlatformUSD            decimal.Decimal `json:"cny_per_platform_usd"`
	USDToCNY                     decimal.Decimal `json:"usd_to_cny"`
	PromotionTrackingEnabled     bool            `json:"promotion_tracking_enabled"`
	PromotionAttributionEnabled  bool            `json:"promotion_attribution_enabled"`
	PromotionAttributionDays     int             `json:"promotion_attribution_days"`
	PromotionAttributionModel    string          `json:"promotion_attribution_model"`
	PromotionCollectSource       bool            `json:"promotion_collect_source"`
	PromotionCollectDevice       bool            `json:"promotion_collect_device"`
	PromotionBotFilterEnabled    bool            `json:"promotion_bot_filter_enabled"`
	PromotionDetailRetentionDays int             `json:"promotion_detail_retention_days"`
}

type DistributionPromotionVisitInput struct {
	PromotionCode string `json:"promotion_code"`
	VisitorToken  string `json:"-"`
	ClientIP      string `json:"-"`
	UserAgent     string `json:"-"`
	LandingPath   string `json:"landing_path"`
	Referrer      string `json:"referrer"`
	UTMSource     string `json:"utm_source"`
	UTMMedium     string `json:"utm_medium"`
	UTMCampaign   string `json:"utm_campaign"`
}

type DistributionPromotionVisitResult struct {
	Tracked         bool   `json:"tracked"`
	Deduplicated    bool   `json:"deduplicated,omitempty"`
	VisitorToken    string `json:"-"`
	AttributionDays int    `json:"attribution_days"`
}

type distributionRegistrationContextKey struct{}
type distributionVisitorTokenContextKey struct{}

func WithDistributionVisitorToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, distributionVisitorTokenContextKey{}, strings.TrimSpace(token))
}

func distributionVisitorTokenFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(distributionVisitorTokenContextKey{}).(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func DistributionVisitorTokenFromContext(ctx context.Context) string {
	return distributionVisitorTokenFromContext(ctx)
}

func WithDistributionRegistrationAttribution(ctx context.Context, value DistributionRegistrationAttribution) context.Context {
	return context.WithValue(ctx, distributionRegistrationContextKey{}, value)
}

func DistributionRegistrationAttributionFromContext(ctx context.Context) (DistributionRegistrationAttribution, bool) {
	value, ok := ctx.Value(distributionRegistrationContextKey{}).(DistributionRegistrationAttribution)
	return value, ok
}

type DistributionRegistrationAttribution struct {
	Code               string
	Type               string
	VisitorTokenHash   string
	FirstVisitID       *int64
	LastVisitID        *int64
	AttributionVisitID *int64
	AttributionSource  string
	AttributionModel   string
}

type DistributionPromotionStatsFilter struct {
	AgentID         int64
	From            time.Time
	To              time.Time
	Page            int
	PageSize        int
	Source          string
	Device          string
	AttributionType string
}

type DistributionPromotionSummary struct {
	TotalVisits            int64   `json:"total_visits"`
	UniqueVisitors         int64   `json:"unique_visitors"`
	BotVisits              int64   `json:"bot_visits"`
	ConvertedVisitors      int64   `json:"converted_visitors"`
	TrackedRegistrations   int64   `json:"tracked_registrations"`
	Registrations          int64   `json:"registrations"`
	DirectRegistrations    int64   `json:"direct_registrations"`
	PersistedRegistrations int64   `json:"persisted_registrations"`
	UntrackedDirect        int64   `json:"untracked_direct"`
	ConversionRate         float64 `json:"conversion_rate"`
}

type DistributionPromotionDailyStat struct {
	Date          string `json:"date"`
	Visits        int64  `json:"visits"`
	Visitors      int64  `json:"visitors"`
	Conversions   int64  `json:"conversions"`
	Registrations int64  `json:"registrations"`
}

type DistributionPromotionSourceStat struct {
	Source        string `json:"source"`
	Visits        int64  `json:"visits"`
	Conversions   int64  `json:"conversions"`
	Registrations int64  `json:"registrations"`
}

type DistributionPromotionAnalyticsMeta struct {
	Cohort              string    `json:"cohort"`
	BotFilterEnabled    bool      `json:"bot_filter_enabled"`
	TrackingEnabled     bool      `json:"tracking_enabled"`
	AttributionEnabled  bool      `json:"attribution_enabled"`
	AttributionDays     int       `json:"attribution_days"`
	AttributionModel    string    `json:"attribution_model"`
	DetailRetentionDays int       `json:"detail_retention_days"`
	RawRetentionDays    int       `json:"raw_retention_days"`
	Timezone            string    `json:"timezone"`
	GeneratedAt         time.Time `json:"generated_at"`
}

type DistributionPromotionAnalytics struct {
	Summary DistributionPromotionSummary       `json:"summary"`
	Daily   []DistributionPromotionDailyStat   `json:"daily"`
	Sources []DistributionPromotionSourceStat  `json:"sources"`
	Meta    DistributionPromotionAnalyticsMeta `json:"meta"`
}

type DistributionPromotionCleanupResult struct {
	ScrubbedRows int
	DeletedRows  int
	SkippedDays  int
}

type DistributionPromotionPrivacyScrubResult struct {
	ScrubbedRows        int
	ExpiredAttributions int
}

type DistributionPromotionVisit struct {
	ID              int64      `json:"id"`
	AgentID         int64      `json:"agent_id"`
	AgentEmail      string     `json:"agent_email,omitempty"`
	PromotionCode   string     `json:"promotion_code"`
	LandingPath     string     `json:"landing_path"`
	Source          string     `json:"source"`
	DeviceType      string     `json:"device_type"`
	IsBot           bool       `json:"is_bot"`
	VisitedAt       time.Time  `json:"visited_at"`
	RegisteredAt    *time.Time `json:"registered_at,omitempty"`
	AttributionType string     `json:"attribution_type,omitempty"`
}

type DistributionGrantAgentInput struct {
	UserID          int64  `json:"user_id"`
	Depth           int    `json:"depth"`
	ParentAgentID   *int64 `json:"parent_agent_id,omitempty"`
	RateOverrideBPS *int   `json:"rate_override_bps,omitempty"`
	PromotionCode   string `json:"promotion_code,omitempty"`
	GrantedBy       int64  `json:"-"`
}

type DistributionGrantChildAgentInput struct {
	Email           string `json:"email"`
	RateOverrideBPS *int   `json:"rate_override_bps,omitempty"`
	PromotionCode   string `json:"promotion_code,omitempty"`
}

type DistributionPayoutAccount struct {
	AlipayName    string `json:"alipay_name"`
	AlipayAccount string `json:"alipay_account"`
}
type DistributionWithdrawal struct {
	ID               int64           `json:"id"`
	RequestNo        string          `json:"request_no"`
	AgentID          int64           `json:"agent_id"`
	AgentUserID      int64           `json:"agent_user_id"`
	AgentEmail       string          `json:"agent_email,omitempty"`
	AgentUsername    string          `json:"agent_username,omitempty"`
	AlipayName       string          `json:"alipay_name"`
	AlipayAccount    string          `json:"alipay_account"`
	AmountCNY        decimal.Decimal `json:"amount_cny"`
	FeeCNY           decimal.Decimal `json:"fee_cny"`
	PayoutCNY        decimal.Decimal `json:"payout_cny"`
	Status           string          `json:"status"`
	ReviewNote       string          `json:"review_note,omitempty"`
	ReviewedBy       *int64          `json:"reviewed_by,omitempty"`
	ReviewerEmail    string          `json:"reviewer_email,omitempty"`
	ReviewedAt       *time.Time      `json:"reviewed_at,omitempty"`
	ApprovedBy       *int64          `json:"approved_by,omitempty"`
	ApproverEmail    string          `json:"approver_email,omitempty"`
	ApprovedAt       *time.Time      `json:"approved_at,omitempty"`
	PaidAt           *time.Time      `json:"paid_at,omitempty"`
	PaymentReference string          `json:"payment_reference,omitempty"`
	AttachmentCount  int64           `json:"attachment_count"`
	CreatedAt        time.Time       `json:"created_at"`
}

type DistributionWithdrawalAttachment struct {
	ID           int64     `json:"id"`
	WithdrawalID int64     `json:"withdrawal_id"`
	ObjectKey    string    `json:"-"`
	OriginalName string    `json:"original_name"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	SHA256       string    `json:"sha256"`
	EvidenceType string    `json:"evidence_type"`
	Note         string    `json:"note"`
	UploadedBy   *int64    `json:"uploaded_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type DistributionWithdrawalEvent struct {
	ID               int64     `json:"id"`
	WithdrawalID     int64     `json:"withdrawal_id"`
	FromStatus       string    `json:"from_status,omitempty"`
	ToStatus         string    `json:"to_status"`
	Note             string    `json:"note"`
	PaymentReference string    `json:"payment_reference,omitempty"`
	ActorUserID      *int64    `json:"actor_user_id,omitempty"`
	ActorEmail       string    `json:"actor_email,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type DistributionWithdrawalDetail struct {
	Withdrawal  *DistributionWithdrawal            `json:"withdrawal"`
	Attachments []DistributionWithdrawalAttachment `json:"attachments"`
	Events      []DistributionWithdrawalEvent      `json:"events"`
}

type DistributionAdminAnomaly struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Severity    string          `json:"severity"`
	EntityType  string          `json:"entity_type"`
	EntityID    int64           `json:"entity_id"`
	Reference   string          `json:"reference"`
	AgentID     *int64          `json:"agent_id,omitempty"`
	AgentEmail  string          `json:"agent_email,omitempty"`
	Subject     string          `json:"subject"`
	AmountCNY   decimal.Decimal `json:"amount_cny"`
	DetectedAt  time.Time       `json:"detected_at"`
	Description string          `json:"description"`
}

type DistributionBatchWithdrawalReviewInput struct {
	WithdrawalIDs []int64 `json:"withdrawal_ids"`
	Status        string  `json:"status"`
	Note          string  `json:"note"`
}

type DistributionBatchWithdrawalReviewResult struct {
	Updated int `json:"updated"`
}
type DistributionCustomer struct {
	UserID             int64           `json:"user_id"`
	Email              string          `json:"email"`
	Username           string          `json:"username"`
	AgentID            int64           `json:"agent_id,omitempty"`
	AgentUserID        int64           `json:"agent_user_id,omitempty"`
	AgentPromotionCode string          `json:"agent_promotion_code,omitempty"`
	AgentEmail         string          `json:"agent_email,omitempty"`
	AgentUsername      string          `json:"agent_username,omitempty"`
	AgentDepth         int             `json:"agent_depth,omitempty"`
	BoundAt            time.Time       `json:"bound_at"`
	OrderCount         int64           `json:"order_count"`
	TotalPaidCNY       decimal.Decimal `json:"total_paid_cny"`
	CommissionCNY      decimal.Decimal `json:"commission_cny"`
	RefundedCNY        decimal.Decimal `json:"refunded_cny"`
	LastPaidAt         *time.Time      `json:"last_paid_at,omitempty"`
}
type DistributionCommission struct {
	ID                 int64           `json:"id"`
	CustomerUserID     int64           `json:"customer_user_id"`
	EntryType          string          `json:"entry_type"`
	RateBPS            int             `json:"rate_bps"`
	OriginalAmountCNY  decimal.Decimal `json:"original_amount_cny"`
	ReversedAmountCNY  decimal.Decimal `json:"reversed_amount_cny"`
	Status             string          `json:"status"`
	CreatedAt          time.Time       `json:"created_at"`
	AvailableAt        time.Time       `json:"available_at"`
	CommissionBaseCNY  decimal.Decimal `json:"commission_base_cny"`
	CustomerEmail      string          `json:"customer_email"`
	CustomerUsername   string          `json:"customer_username"`
	PaymentType        string          `json:"payment_type"`
	PaymentOrderID     int64           `json:"payment_order_id"`
	OrderNo            string          `json:"order_no"`
	BeneficiaryAgentID int64           `json:"beneficiary_agent_id"`
	AgentEmail         string          `json:"agent_email"`
	AgentUsername      string          `json:"agent_username"`
	PaymentCurrency    string          `json:"payment_currency"`
	ActualPaidAmount   decimal.Decimal `json:"actual_paid_amount"`
	FXRateToCNY        decimal.Decimal `json:"fx_rate_to_cny"`
}

type DistributionSettlementRules struct {
	WithdrawalEnabled         bool            `json:"withdrawal_enabled"`
	MinimumWithdrawalCNY      decimal.Decimal `json:"minimum_withdrawal_cny"`
	MaximumWithdrawalCNY      decimal.Decimal `json:"maximum_withdrawal_cny"`
	WithdrawalFeeRateBPS      int             `json:"withdrawal_fee_rate_bps"`
	WithdrawalFeeFixedCNY     decimal.Decimal `json:"withdrawal_fee_fixed_cny"`
	DailyWithdrawalLimitCNY   decimal.Decimal `json:"daily_withdrawal_limit_cny"`
	MonthlyWithdrawalLimitCNY decimal.Decimal `json:"monthly_withdrawal_limit_cny"`
	CNYPerPlatformUSD         decimal.Decimal `json:"cny_per_platform_usd"`
	FreezeHours               int             `json:"freeze_hours"`
}
type DistributionRepository interface {
	ValidatePromotionCode(ctx context.Context, promotionCode string) error
	BindCustomerByCode(ctx context.Context, userID int64, promotionCode string) error
	TrackPromotionVisit(ctx context.Context, input DistributionPromotionVisitInput, visitorTokenHash, ipHash, deviceType string, isBot bool) (*DistributionPromotionVisitResult, error)
	ResolvePromotionAttribution(ctx context.Context, visitorTokenHash string, promotionCode ...string) (*DistributionRegistrationAttribution, error)
	GetPromotionAnalytics(ctx context.Context, userID int64, filter DistributionPromotionStatsFilter, admin bool) (*DistributionPromotionAnalytics, error)
	ListPromotionVisits(ctx context.Context, userID int64, filter DistributionPromotionStatsFilter, admin bool) ([]DistributionPromotionVisit, int64, error)
	ScrubPromotionDetails(ctx context.Context, batchSize int) (DistributionPromotionPrivacyScrubResult, error)
	CleanupPromotionDetails(ctx context.Context, batchSize int) (DistributionPromotionCleanupResult, error)
	GetAgentByUserID(ctx context.Context, userID int64) (*DistributionAgent, error)
	IsAgent(ctx context.Context, userID int64) (bool, error)
	GetOverview(ctx context.Context, userID int64) (*DistributionOverview, error)
	AccruePaidOrder(ctx context.Context, input DistributionCommissionInput) (*DistributionCommissionResult, error)
	ReverseRefund(ctx context.Context, paymentOrderID int64, refundedActualAmount decimal.Decimal) error
	AdminGetSettings(ctx context.Context) (*DistributionSettings, error)
	AdminGetOverview(ctx context.Context) (*DistributionAdminOverview, error)
	AdminUpdateSettings(ctx context.Context, settings DistributionSettings, adminID int64) error
	AdminSetFXRate(ctx context.Context, currency string, rate decimal.Decimal, adminID int64) error
	AdminGrantAgent(ctx context.Context, input DistributionGrantAgentInput) (*DistributionAgent, error)
	AdminUpdateAgentStatus(ctx context.Context, agentID, adminID int64, status, reason string) error
	AdminUpdateAgentRate(ctx context.Context, agentID, adminID int64, rateOverrideBPS *int, reason string) error
	AdminUpdateAgentRecruitmentPermission(ctx context.Context, agentID, adminID int64, enabled bool, reason string) error
	AdminUpdateAgentPromotionStatsPermission(ctx context.Context, agentID, adminID int64, enabled bool, reason string) error
	AdminReviewWithdrawal(ctx context.Context, withdrawalID, adminID int64, status, note, reference string) error
	AdminBatchReviewWithdrawals(ctx context.Context, withdrawalIDs []int64, adminID int64, status, note string) error
	AdminListWithdrawals(ctx context.Context, filter DistributionAdminWithdrawalListFilter) ([]DistributionWithdrawal, int64, error)
	AdminGetWithdrawal(ctx context.Context, withdrawalID int64) (*DistributionWithdrawalDetail, error)
	AdminCreateWithdrawalAttachments(ctx context.Context, withdrawalID, adminID int64, attachments []DistributionWithdrawalAttachment) error
	GetPayoutAccount(ctx context.Context, userID int64) (*DistributionPayoutAccount, error)
	UpsertPayoutAccount(ctx context.Context, userID int64, account DistributionPayoutAccount) error
	RequestWithdrawal(ctx context.Context, userID int64, amount decimal.Decimal) (*DistributionWithdrawal, error)
	ListWithdrawals(ctx context.Context, userID int64, filter DistributionUserListFilter) ([]DistributionWithdrawal, int64, error)
	ConvertToBalance(ctx context.Context, userID int64, amount decimal.Decimal) (decimal.Decimal, error)
	ListCustomers(ctx context.Context, userID int64, filter DistributionUserListFilter) ([]DistributionCustomer, int64, error)
	ListCommissions(ctx context.Context, userID int64, filter DistributionUserListFilter) ([]DistributionCommission, int64, error)
	ListTeam(ctx context.Context, userID int64, filter DistributionUserListFilter) ([]DistributionAgent, int64, error)
	UpdateTeamAgentStatus(ctx context.Context, userID, agentID int64, status string) error
	AdminUpdateLevel(ctx context.Context, depth, defaultRateBPS, maxChildRateBPS int, active bool) error
	AdminCorrectCustomerBinding(ctx context.Context, userID, agentID, adminID int64, reason string) error
	AdminListAgents(ctx context.Context, filter DistributionAdminListFilter) ([]DistributionAgent, int64, error)
	AdminListCustomers(ctx context.Context, filter DistributionAdminListFilter) ([]DistributionCustomer, int64, error)
	AdminLookupAgentCandidates(ctx context.Context, query string) ([]DistributionUserOption, error)
	LookupEligibleUserByExactEmail(ctx context.Context, email string) (int64, error)
	AdminLookupAgents(ctx context.Context, query string, includeInactive bool) ([]DistributionAgentOption, error)
	AdminListAgentEvents(ctx context.Context, agentID int64) ([]DistributionAgentEvent, error)
	AdminListBindingEvents(ctx context.Context, customerUserID int64) ([]DistributionBindingEvent, error)
	AdminListCommissions(ctx context.Context, filter DistributionAdminCommissionListFilter) ([]DistributionCommission, int64, error)
	AdminListAnomalies(ctx context.Context, filter DistributionAdminAnomalyListFilter) ([]DistributionAdminAnomaly, int64, error)
	ReleaseMaturedCommissions(ctx context.Context) (int, error)
}

func (s *DistributionService) GetPayoutAccount(ctx context.Context, userID int64) (*DistributionPayoutAccount, error) {
	return s.repo.GetPayoutAccount(ctx, userID)
}

func (s *DistributionService) AdminGetOverview(ctx context.Context) (*DistributionAdminOverview, error) {
	return s.repo.AdminGetOverview(ctx)
}

func (s *DistributionService) AdminListAgentEvents(ctx context.Context, agentID int64) ([]DistributionAgentEvent, error) {
	if agentID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_AGENT", "invalid agent")
	}
	return s.repo.AdminListAgentEvents(ctx, agentID)
}

func (s *DistributionService) AdminListBindingEvents(ctx context.Context, customerUserID int64) ([]DistributionBindingEvent, error) {
	if customerUserID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_CUSTOMER", "invalid customer")
	}
	return s.repo.AdminListBindingEvents(ctx, customerUserID)
}
func (s *DistributionService) UpsertPayoutAccount(ctx context.Context, userID int64, a DistributionPayoutAccount) error {
	a.AlipayName = strings.TrimSpace(a.AlipayName)
	a.AlipayAccount = strings.TrimSpace(a.AlipayAccount)
	if a.AlipayName == "" || a.AlipayAccount == "" {
		return infraerrors.BadRequest("INVALID_PAYOUT_ACCOUNT", "Alipay name and account are required")
	}
	return s.repo.UpsertPayoutAccount(ctx, userID, a)
}
func (s *DistributionService) RequestWithdrawal(ctx context.Context, userID int64, amount decimal.Decimal) (*DistributionWithdrawal, error) {
	if !amount.IsPositive() {
		return nil, infraerrors.BadRequest("INVALID_WITHDRAWAL_AMOUNT", "invalid withdrawal amount")
	}
	return s.repo.RequestWithdrawal(ctx, userID, amount.Round(8))
}
func (s *DistributionService) ListWithdrawals(ctx context.Context, userID int64, filter DistributionUserListFilter) ([]DistributionWithdrawal, int64, error) {
	return s.repo.ListWithdrawals(ctx, userID, normalizeDistributionUserListFilter(filter))
}
func (s *DistributionService) ConvertToBalance(ctx context.Context, userID int64, amount decimal.Decimal) (decimal.Decimal, error) {
	if !amount.IsPositive() {
		return decimal.Zero, infraerrors.BadRequest("INVALID_CONVERSION_AMOUNT", "invalid conversion amount")
	}
	return s.repo.ConvertToBalance(ctx, userID, amount.Round(8))
}
func (s *DistributionService) ListCustomers(ctx context.Context, userID int64, filter DistributionUserListFilter) ([]DistributionCustomer, int64, error) {
	return s.repo.ListCustomers(ctx, userID, normalizeDistributionUserListFilter(filter))
}
func (s *DistributionService) ListCommissions(ctx context.Context, userID int64, filter DistributionUserListFilter) ([]DistributionCommission, int64, error) {
	return s.repo.ListCommissions(ctx, userID, normalizeDistributionUserListFilter(filter))
}
func (s *DistributionService) ListTeam(ctx context.Context, userID int64, filter DistributionUserListFilter) ([]DistributionAgent, int64, error) {
	return s.repo.ListTeam(ctx, userID, normalizeDistributionUserListFilter(filter))
}

func (s *DistributionService) UpdateTeamAgentStatus(ctx context.Context, userID, agentID int64, status string) error {
	if agentID <= 0 || (status != "active" && status != "suspended") {
		return infraerrors.BadRequest("INVALID_AGENT_STATUS", "invalid team agent status")
	}
	return s.repo.UpdateTeamAgentStatus(ctx, userID, agentID, status)
}

func normalizeDistributionUserListFilter(filter DistributionUserListFilter) DistributionUserListFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Status != "active" && filter.Status != "suspended" && filter.Status != "revoked" &&
		filter.Status != "frozen" && filter.Status != "available" && filter.Status != "reserved" &&
		filter.Status != "paid" && filter.Status != "converted" && filter.Status != "reversed" &&
		filter.Status != "pending" && filter.Status != "approved" && filter.Status != "rejected" &&
		filter.Status != "paying" && filter.Status != "failed" && filter.Status != "cancelled" {
		filter.Status = ""
	}
	if filter.EntryType != "direct" && filter.EntryType != "team" {
		filter.EntryType = ""
	}
	return filter
}

func (s *DistributionService) GetSettlementRules(ctx context.Context) (*DistributionSettlementRules, error) {
	settings, err := s.repo.AdminGetSettings(ctx)
	if err != nil {
		return nil, err
	}
	return &DistributionSettlementRules{
		WithdrawalEnabled: settings.WithdrawalEnabled, MinimumWithdrawalCNY: settings.MinimumWithdrawalCNY,
		MaximumWithdrawalCNY: settings.MaximumWithdrawalCNY, WithdrawalFeeRateBPS: settings.WithdrawalFeeRateBPS,
		WithdrawalFeeFixedCNY: settings.WithdrawalFeeFixedCNY, DailyWithdrawalLimitCNY: settings.DailyWithdrawalLimitCNY,
		MonthlyWithdrawalLimitCNY: settings.MonthlyWithdrawalLimitCNY, CNYPerPlatformUSD: settings.CNYPerPlatformUSD,
		FreezeHours: settings.FreezeHours,
	}, nil
}

func (s *DistributionService) AdminUpdateLevel(ctx context.Context, depth, defaultRateBPS, maxChildRateBPS int, active bool) error {
	if (depth != 1 && depth != 2) || defaultRateBPS < 0 || defaultRateBPS > 10000 || maxChildRateBPS < 0 || maxChildRateBPS > defaultRateBPS {
		return infraerrors.BadRequest("INVALID_AGENT_LEVEL", "invalid agent level")
	}
	return s.repo.AdminUpdateLevel(ctx, depth, defaultRateBPS, maxChildRateBPS, active)
}

func (s *DistributionService) AdminUpdateAgentPromotionStatsPermission(ctx context.Context, agentID, adminID int64, enabled bool, reason string) error {
	if agentID <= 0 || strings.TrimSpace(reason) == "" {
		return infraerrors.BadRequest("INVALID_PROMOTION_STATS_PERMISSION", "agent and reason are required")
	}
	return s.repo.AdminUpdateAgentPromotionStatsPermission(ctx, agentID, adminID, enabled, strings.TrimSpace(reason))
}
func (s *DistributionService) AdminCorrectCustomerBinding(ctx context.Context, userID, agentID, adminID int64, reason string) error {
	if userID <= 0 || agentID <= 0 || strings.TrimSpace(reason) == "" {
		return infraerrors.BadRequest("INVALID_BINDING_CORRECTION", "user, agent and reason are required")
	}
	return s.repo.AdminCorrectCustomerBinding(ctx, userID, agentID, adminID, strings.TrimSpace(reason))
}
func normalizeDistributionAdminListFilter(filter DistributionAdminListFilter) DistributionAdminListFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Status != "active" && filter.Status != "suspended" && filter.Status != "revoked" {
		filter.Status = ""
	}
	if filter.Depth != 1 && filter.Depth != 2 {
		filter.Depth = 0
	}
	if filter.AgentID < 0 {
		filter.AgentID = 0
	}
	filter.SortOrder = normalizeDistributionSortOrder(filter.SortOrder)
	return filter
}

func normalizeDistributionSortOrder(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "asc") {
		return "asc"
	}
	return "desc"
}

func normalizeAdminCommissionFilter(filter DistributionAdminCommissionListFilter) DistributionAdminCommissionListFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.EntryType != "direct" && filter.EntryType != "team" {
		filter.EntryType = ""
	}
	if filter.Status != "frozen" && filter.Status != "available" && filter.Status != "reserved" &&
		filter.Status != "paid" && filter.Status != "converted" && filter.Status != "reversed" {
		filter.Status = ""
	}
	filter.PaymentType = strings.TrimSpace(filter.PaymentType)
	if filter.AgentID < 0 {
		filter.AgentID = 0
	}
	filter.SortOrder = normalizeDistributionSortOrder(filter.SortOrder)
	return filter
}

func normalizeAdminWithdrawalFilter(filter DistributionAdminWithdrawalListFilter) DistributionAdminWithdrawalListFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Status != "pending" && filter.Status != "approved" && filter.Status != "paying" &&
		filter.Status != "paid" && filter.Status != "rejected" && filter.Status != "failed" && filter.Status != "cancelled" {
		filter.Status = ""
	}
	if filter.AgentID < 0 {
		filter.AgentID = 0
	}
	filter.SortOrder = normalizeDistributionSortOrder(filter.SortOrder)
	return filter
}

func normalizeAdminAnomalyFilter(filter DistributionAdminAnomalyListFilter) DistributionAdminAnomalyListFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	filter.Search = strings.TrimSpace(filter.Search)
	switch filter.Type {
	case "overdue_withdrawal", "pending_fx", "agent_debt", "inactive_agent_customers", "matured_commission":
	default:
		filter.Type = ""
	}
	switch filter.Severity {
	case "critical", "high", "medium":
	default:
		filter.Severity = ""
	}
	filter.SortOrder = normalizeDistributionSortOrder(filter.SortOrder)
	return filter
}

func (s *DistributionService) AdminListAgents(ctx context.Context, filter DistributionAdminListFilter) ([]DistributionAgent, int64, error) {
	return s.repo.AdminListAgents(ctx, normalizeDistributionAdminListFilter(filter))
}
func (s *DistributionService) AdminListCustomers(ctx context.Context, filter DistributionAdminListFilter) ([]DistributionCustomer, int64, error) {
	return s.repo.AdminListCustomers(ctx, normalizeDistributionAdminListFilter(filter))
}
func (s *DistributionService) AdminLookupAgentCandidates(ctx context.Context, query string) ([]DistributionUserOption, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []DistributionUserOption{}, nil
	}
	return s.repo.AdminLookupAgentCandidates(ctx, query)
}
func (s *DistributionService) AdminLookupAgents(ctx context.Context, query string, includeInactive ...bool) ([]DistributionAgentOption, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []DistributionAgentOption{}, nil
	}
	return s.repo.AdminLookupAgents(ctx, query, len(includeInactive) > 0 && includeInactive[0])
}
func (s *DistributionService) AdminListCommissions(ctx context.Context, filter DistributionAdminCommissionListFilter) ([]DistributionCommission, int64, error) {
	return s.repo.AdminListCommissions(ctx, normalizeAdminCommissionFilter(filter))
}
func (s *DistributionService) AdminListAnomalies(ctx context.Context, filter DistributionAdminAnomalyListFilter) ([]DistributionAdminAnomaly, int64, error) {
	return s.repo.AdminListAnomalies(ctx, normalizeAdminAnomalyFilter(filter))
}

type DistributionService struct {
	repo                DistributionRepository
	evidenceStorage     TicketStorage
	evidenceConfig      config.TicketStorageConfig
	maturity            *distributionMaturityRuntime
	promotionCleanup    *distributionPromotionCleanupRuntime
	trackingHashSecrets []string
}

func NewDistributionService(repo DistributionRepository, trackingHashSecrets ...string) *DistributionService {
	return &DistributionService{repo: repo, trackingHashSecrets: normalizeTrackingHashSecrets(trackingHashSecrets)}
}

func ProvideDistributionService(repo DistributionRepository, storage TicketStorage, cfg *config.Config, lockCache LeaderLockCache, db *sql.DB) *DistributionService {
	secrets := normalizeTrackingHashSecrets(cfg.DistributionTracking.HashSecrets)
	if len(secrets) == 0 {
		slog.Error("distribution tracking disabled: an independent distribution_tracking.hash_secrets value is required")
	}
	svc := &DistributionService{repo: repo, evidenceStorage: storage, evidenceConfig: cfg.TicketStorage, trackingHashSecrets: secrets}
	if db != nil {
		svc.maturity = newDistributionMaturityRuntime(repo, lockCache, db, time.Minute)
		svc.maturity.Start()
		svc.promotionCleanup = newDistributionPromotionCleanupRuntime(
			repo,
			lockCache,
			db,
			time.Duration(cfg.DistributionTracking.CleanupIntervalMinutes)*time.Minute,
			cfg.DistributionTracking.CleanupBatchSize,
			cfg.DistributionTracking.CleanupEnabled,
		)
		svc.promotionCleanup.Start()
		if !cfg.DistributionTracking.CleanupEnabled {
			slog.Info("distribution promotion physical cleanup disabled; archived visit rows will not be deleted")
		}
	}
	return svc
}

func normalizeDistributionCode(raw string) string { return strings.ToUpper(strings.TrimSpace(raw)) }

func validDistributionCode(code string) bool {
	if len(code) < 4 || len(code) > 32 {
		return false
	}
	for i := 0; i < len(code); i++ {
		c := code[i]
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			continue
		}
		return false
	}
	return true
}

func (s *DistributionService) BindCustomerByCode(ctx context.Context, userID int64, rawCode string) error {
	code := normalizeDistributionCode(rawCode)
	if code == "" {
		return nil
	}
	if s == nil || s.repo == nil {
		return infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "distribution service unavailable")
	}
	if userID <= 0 || !validDistributionCode(code) {
		return ErrDistributionCodeInvalid
	}
	return s.repo.BindCustomerByCode(ctx, userID, code)
}

func (s *DistributionService) ResolveRegistrationAttribution(ctx context.Context, explicitCode, affiliateCode string) (string, error) {
	explicitCode = normalizeDistributionCode(explicitCode)
	if explicitCode != "" {
		if err := s.ValidatePromotionCode(ctx, explicitCode); err != nil {
			return "", err
		}
		return explicitCode, nil
	}
	if strings.TrimSpace(affiliateCode) != "" || s == nil || s.repo == nil {
		return "", nil
	}
	token := distributionVisitorTokenFromContext(ctx)
	if token == "" {
		return "", nil
	}
	for _, hash := range s.hashVisitorTokenCandidates(token) {
		attr, err := s.repo.ResolvePromotionAttribution(ctx, hash)
		if err != nil {
			return "", err
		}
		if attr != nil {
			return attr.Code, nil
		}
	}
	return "", nil
}

func (s *DistributionService) PrepareRegistrationAttribution(ctx context.Context, explicitCode, affiliateCode string) (context.Context, string, error) {
	code, err := s.ResolveRegistrationAttribution(ctx, explicitCode, affiliateCode)
	if err != nil {
		return ctx, "", err
	}
	if code == "" {
		return ctx, "", nil
	}
	attr := DistributionRegistrationAttribution{Code: code, Type: "direct"}
	hashes := s.hashVisitorTokenCandidates(distributionVisitorTokenFromContext(ctx))
	if len(hashes) > 0 && strings.TrimSpace(affiliateCode) == "" {
		lookupCode := ""
		if strings.TrimSpace(explicitCode) != "" {
			lookupCode = code
		}
		for _, hash := range hashes {
			resolved, resolveErr := s.repo.ResolvePromotionAttribution(ctx, hash, lookupCode)
			if resolveErr != nil {
				return ctx, "", resolveErr
			}
			if resolved != nil {
				attr = *resolved
				break
			}
		}
	}
	return WithDistributionRegistrationAttribution(ctx, attr), code, nil
}

func (s *DistributionService) hashVisitorToken(value string) string {
	candidates := s.hashVisitorTokenCandidates(value)
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0]
}

func normalizeTrackingHashSecrets(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func (s *DistributionService) hashVisitorTokenCandidates(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	secrets := s.trackingHashSecrets
	if len(secrets) == 0 {
		return nil
	}
	out := make([]string, 0, len(secrets))
	for _, secret := range secrets {
		h := sha256.Sum256([]byte(secret + "\n" + value))
		out = append(out, hex.EncodeToString(h[:]))
	}
	return out
}

func (s *DistributionService) TrackPromotionVisit(ctx context.Context, input DistributionPromotionVisitInput) (*DistributionPromotionVisitResult, error) {
	if s == nil || len(s.trackingHashSecrets) == 0 {
		return nil, infraerrors.ServiceUnavailable("PROMOTION_TRACKING_NOT_CONFIGURED", "promotion tracking is not configured")
	}
	input.PromotionCode = normalizeDistributionCode(input.PromotionCode)
	if !validDistributionCode(input.PromotionCode) {
		return nil, ErrDistributionCodeInvalid
	}
	input.LandingPath = normalizeTrackingLandingPath(input.LandingPath)
	input.Referrer = normalizeTrackingReferrerHost(input.Referrer)
	input.UTMSource = strings.TrimSpace(input.UTMSource)
	input.UTMMedium = strings.TrimSpace(input.UTMMedium)
	input.UTMCampaign = strings.TrimSpace(input.UTMCampaign)
	if !validTrackingText(input.LandingPath, 512) || !validTrackingText(input.Referrer, 2048) ||
		!validTrackingText(input.UTMSource, 128) || !validTrackingText(input.UTMMedium, 128) ||
		!validTrackingText(input.UTMCampaign, 128) {
		return nil, ErrPromotionTrackingInput
	}
	if input.VisitorToken == "" {
		input.VisitorToken = uuid.NewString()
	}
	if input.LandingPath == "" {
		input.LandingPath = "/register"
	}
	if !validTrackingText(input.Referrer, 255) {
		return nil, ErrPromotionTrackingInput
	}
	device := "desktop"
	lowerUA := strings.ToLower(input.UserAgent)
	isBot := strings.Contains(lowerUA, "bot") || strings.Contains(lowerUA, "crawler") || strings.Contains(lowerUA, "spider") || strings.Contains(lowerUA, "headless")
	if isBot {
		device = "bot"
	} else if strings.Contains(lowerUA, "ipad") || strings.Contains(lowerUA, "tablet") {
		device = "tablet"
	} else if strings.Contains(lowerUA, "mobile") || strings.Contains(lowerUA, "android") || strings.Contains(lowerUA, "iphone") {
		device = "mobile"
	}
	result, err := s.repo.TrackPromotionVisit(ctx, input, s.hashVisitorToken(input.VisitorToken), s.hashVisitorToken(input.ClientIP), device, isBot)
	if result != nil {
		result.VisitorToken = input.VisitorToken
	}
	return result, err
}

func normalizeTrackingLandingPath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Path == "" || !strings.HasPrefix(parsed.Path, "/") {
		return ""
	}
	return parsed.Path
}

func normalizeTrackingReferrerHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" {
		return ""
	}
	return strings.ToLower(parsed.Hostname())
}

type promotionTrackingStatusRepository interface {
	PromotionTrackingEnabled(ctx context.Context) (bool, error)
}

func (s *DistributionService) PromotionTrackingEnabled(ctx context.Context) (bool, error) {
	if s == nil || s.repo == nil {
		return false, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "distribution service unavailable")
	}
	if len(s.trackingHashSecrets) == 0 {
		return false, nil
	}
	repo, ok := s.repo.(promotionTrackingStatusRepository)
	if !ok {
		return false, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "promotion tracking status unavailable")
	}
	return repo.PromotionTrackingEnabled(ctx)
}

func validTrackingText(value string, maxRunes int) bool {
	return utf8.ValidString(value) && utf8.RuneCountInString(value) <= maxRunes && !strings.ContainsRune(value, '\x00')
}

func (s *DistributionService) GetPromotionAnalytics(ctx context.Context, userID int64, filter DistributionPromotionStatsFilter, admin bool) (*DistributionPromotionAnalytics, error) {
	if err := normalizePromotionStatsFilter(&filter, false); err != nil {
		return nil, err
	}
	return s.repo.GetPromotionAnalytics(ctx, userID, filter, admin)
}

func (s *DistributionService) ListPromotionVisits(ctx context.Context, userID int64, filter DistributionPromotionStatsFilter, admin bool) ([]DistributionPromotionVisit, int64, error) {
	if err := normalizePromotionStatsFilter(&filter, true); err != nil {
		return nil, 0, err
	}
	return s.repo.ListPromotionVisits(ctx, userID, filter, admin)
}

func normalizePromotionStatsFilter(filter *DistributionPromotionStatsFilter, paginate bool) error {
	now := time.Now()
	if filter.From.IsZero() {
		filter.From = now.AddDate(0, 0, -30)
	}
	if filter.To.IsZero() {
		filter.To = now
	}
	filter.Source = strings.TrimSpace(filter.Source)
	filter.Device = strings.TrimSpace(filter.Device)
	filter.AttributionType = strings.TrimSpace(filter.AttributionType)
	validDevice := filter.Device == "" || filter.Device == "desktop" || filter.Device == "mobile" || filter.Device == "tablet" || filter.Device == "bot" || filter.Device == "unknown"
	validAttribution := filter.AttributionType == "" || filter.AttributionType == "direct" || filter.AttributionType == "persisted" || filter.AttributionType == "unregistered"
	if !filter.To.After(filter.From) || filter.To.Sub(filter.From) > 366*24*time.Hour || !validTrackingText(filter.Source, 255) || !validDevice || !validAttribution {
		return ErrPromotionTrackingInput
	}
	if !paginate {
		return nil
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	const maxPromotionVisitOffset = 10_000_000
	if filter.Page > 1+maxPromotionVisitOffset/filter.PageSize {
		return ErrPromotionTrackingInput
	}
	return nil
}

func (s *DistributionService) ValidatePromotionCode(ctx context.Context, rawCode string) error {
	code := normalizeDistributionCode(rawCode)
	if !validDistributionCode(code) {
		return ErrDistributionCodeInvalid
	}
	if s == nil || s.repo == nil {
		return infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "distribution service unavailable")
	}
	return s.repo.ValidatePromotionCode(ctx, code)
}

func (s *DistributionService) GetOverview(ctx context.Context, userID int64) (*DistributionOverview, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "distribution service unavailable")
	}
	return s.repo.GetOverview(ctx, userID)
}

func (s *DistributionService) IsAgent(ctx context.Context, userID int64) bool {
	if s == nil || s.repo == nil {
		return false
	}
	ok, err := s.repo.IsAgent(ctx, userID)
	return err == nil && ok
}

type DistributionAccess struct {
	Enabled               bool   `json:"enabled"`
	IsAgent               bool   `json:"is_agent"`
	Depth                 int    `json:"depth,omitempty"`
	Status                string `json:"status,omitempty"`
	CanRecruitSubagents   bool   `json:"can_recruit_subagents"`
	CanViewPromotionStats bool   `json:"can_view_promotion_stats"`
}

func (s *DistributionService) GetAccess(ctx context.Context, userID int64) (*DistributionAccess, error) {
	settings, err := s.repo.AdminGetSettings(ctx)
	if err != nil {
		return nil, err
	}
	isAgent, err := s.repo.IsAgent(ctx, userID)
	if err != nil {
		return nil, err
	}
	access := &DistributionAccess{Enabled: settings.Enabled, IsAgent: isAgent}
	if isAgent {
		agent, agentErr := s.repo.GetAgentByUserID(ctx, userID)
		if agentErr != nil {
			return nil, agentErr
		}
		access.Depth = agent.Depth
		access.Status = agent.Status
		access.CanRecruitSubagents = agent.CanRecruitSubagents
		access.CanViewPromotionStats = agent.CanViewPromotionStats
	}
	return access, nil
}

func (s *DistributionService) EnsureAgentAccess(ctx context.Context, userID int64) error {
	access, err := s.GetAccess(ctx, userID)
	if err != nil {
		return err
	}
	if !access.Enabled {
		return ErrDistributionDisabled
	}
	if !access.IsAgent {
		return ErrDistributionNotAgent
	}
	return nil
}

func (s *DistributionService) AdminGetSettings(ctx context.Context) (*DistributionSettings, error) {
	return s.repo.AdminGetSettings(ctx)
}
func (s *DistributionService) AdminUpdateSettings(ctx context.Context, v DistributionSettings, adminID int64) error {
	if v.L1DefaultRateBPS < 0 || v.L1DefaultRateBPS > 10000 || v.L1MaxChildRateBPS < 0 || v.L1MaxChildRateBPS > v.L1DefaultRateBPS || v.L2DefaultRateBPS < 0 || v.L2DefaultRateBPS > v.L1MaxChildRateBPS || v.FreezeHours < 0 || v.MinimumWithdrawalCNY.IsNegative() || v.MaximumWithdrawalCNY.LessThan(v.MinimumWithdrawalCNY) || v.WithdrawalFeeRateBPS < 0 || v.WithdrawalFeeRateBPS > 10000 || !v.CNYPerPlatformUSD.Equal(decimal.NewFromInt(1)) || !v.USDToCNY.IsPositive() || v.PromotionAttributionDays < 1 || v.PromotionAttributionDays > 365 || (v.PromotionAttributionModel != "first_touch" && v.PromotionAttributionModel != "last_touch") || v.PromotionDetailRetentionDays < 30 || v.PromotionDetailRetentionDays > 730 {
		return infraerrors.BadRequest("INVALID_DISTRIBUTION_SETTINGS", "invalid distribution settings")
	}
	return s.repo.AdminUpdateSettings(ctx, v, adminID)
}
func (s *DistributionService) AdminSetFXRate(ctx context.Context, currency string, rate decimal.Decimal, adminID int64) error {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" || !rate.IsPositive() {
		return infraerrors.BadRequest("INVALID_FX_RATE", "invalid FX rate")
	}
	return s.repo.AdminSetFXRate(ctx, currency, rate, adminID)
}
func (s *DistributionService) AdminGrantAgent(ctx context.Context, input DistributionGrantAgentInput) (*DistributionAgent, error) {
	if input.UserID <= 0 || input.Depth != 1 {
		return nil, infraerrors.BadRequest("INVALID_AGENT", "invalid agent")
	}
	input.ParentAgentID = nil
	input.PromotionCode = normalizeDistributionCode(input.PromotionCode)
	if input.PromotionCode != "" && !validDistributionCode(input.PromotionCode) {
		return nil, ErrDistributionCodeInvalid
	}
	return s.repo.AdminGrantAgent(ctx, input)
}
func (s *DistributionService) GrantL2Agent(ctx context.Context, l1UserID int64, request DistributionGrantChildAgentInput) (*DistributionAgent, error) {
	parent, err := s.repo.GetAgentByUserID(ctx, l1UserID)
	if err != nil {
		return nil, err
	}
	if parent.Depth != 1 || parent.Status != "active" {
		return nil, infraerrors.Forbidden("L1_AGENT_REQUIRED", "active L1 agent required")
	}
	if !parent.CanRecruitSubagents {
		return nil, ErrSubagentRecruitmentDenied
	}
	email := strings.ToLower(strings.TrimSpace(request.Email))
	parsed, parseErr := mail.ParseAddress(email)
	if parseErr != nil || !strings.EqualFold(parsed.Address, email) {
		return nil, ErrSubagentCandidateUnavailable
	}
	userID, lookupErr := s.repo.LookupEligibleUserByExactEmail(ctx, email)
	if lookupErr != nil {
		return nil, ErrSubagentCandidateUnavailable
	}
	input := DistributionGrantAgentInput{UserID: userID, RateOverrideBPS: request.RateOverrideBPS, PromotionCode: request.PromotionCode}
	input.Depth = 2
	input.ParentAgentID = &parent.ID
	input.GrantedBy = l1UserID
	input.PromotionCode = normalizeDistributionCode(input.PromotionCode)
	if input.UserID <= 0 || (input.PromotionCode != "" && !validDistributionCode(input.PromotionCode)) {
		return nil, infraerrors.BadRequest("INVALID_AGENT", "invalid agent")
	}
	return s.repo.AdminGrantAgent(ctx, input)
}

func (s *DistributionService) AdminUpdateAgentRecruitmentPermission(ctx context.Context, agentID, adminID int64, enabled bool, reason string) error {
	reason = strings.TrimSpace(reason)
	if agentID <= 0 || adminID <= 0 || reason == "" {
		return infraerrors.BadRequest("INVALID_RECRUITMENT_PERMISSION", "agent, administrator and reason are required")
	}
	return s.repo.AdminUpdateAgentRecruitmentPermission(ctx, agentID, adminID, enabled, reason)
}
func (s *DistributionService) AdminUpdateAgentStatus(ctx context.Context, agentID, adminID int64, status, reason string) error {
	if status != "active" && status != "suspended" && status != "revoked" {
		return infraerrors.BadRequest("INVALID_AGENT_STATUS", "invalid agent status")
	}
	reason = strings.TrimSpace(reason)
	if agentID <= 0 || adminID <= 0 || reason == "" {
		return infraerrors.BadRequest("INVALID_AGENT_STATUS", "agent, administrator and reason are required")
	}
	return s.repo.AdminUpdateAgentStatus(ctx, agentID, adminID, status, reason)
}
func (s *DistributionService) AdminUpdateAgentRate(ctx context.Context, agentID, adminID int64, rateOverrideBPS *int, reason string) error {
	reason = strings.TrimSpace(reason)
	if agentID <= 0 || adminID <= 0 || reason == "" || (rateOverrideBPS != nil && (*rateOverrideBPS < 0 || *rateOverrideBPS > 10000)) {
		return infraerrors.BadRequest("INVALID_AGENT_RATE", "valid agent, rate and reason are required")
	}
	return s.repo.AdminUpdateAgentRate(ctx, agentID, adminID, rateOverrideBPS, reason)
}
func (s *DistributionService) AdminReviewWithdrawal(ctx context.Context, withdrawalID, adminID int64, status, note, reference string) error {
	if status != "approved" && status != "rejected" && status != "paying" && status != "paid" && status != "failed" {
		return infraerrors.BadRequest("INVALID_WITHDRAWAL_STATUS", "invalid withdrawal status")
	}
	note = strings.TrimSpace(note)
	reference = strings.TrimSpace(reference)
	if withdrawalID <= 0 || adminID <= 0 || (status == "paid" && reference == "") || ((status == "rejected" || status == "failed") && note == "") {
		return infraerrors.BadRequest("INVALID_WITHDRAWAL_REVIEW", "withdrawal transition requires complete audit information")
	}
	return s.repo.AdminReviewWithdrawal(ctx, withdrawalID, adminID, status, note, reference)
}
func (s *DistributionService) AdminBatchReviewWithdrawals(ctx context.Context, input DistributionBatchWithdrawalReviewInput, adminID int64) (*DistributionBatchWithdrawalReviewResult, error) {
	input.Note = strings.TrimSpace(input.Note)
	if adminID <= 0 || len(input.WithdrawalIDs) == 0 || len(input.WithdrawalIDs) > 100 || (input.Status != "approved" && input.Status != "rejected") || (input.Status == "rejected" && input.Note == "") {
		return nil, infraerrors.BadRequest("INVALID_BATCH_WITHDRAWAL_REVIEW", "batch review requires 1-100 withdrawals, a supported status and complete audit information")
	}
	seen := make(map[int64]struct{}, len(input.WithdrawalIDs))
	ids := make([]int64, 0, len(input.WithdrawalIDs))
	for _, id := range input.WithdrawalIDs {
		if id <= 0 {
			return nil, infraerrors.BadRequest("INVALID_BATCH_WITHDRAWAL_REVIEW", "invalid withdrawal id")
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if err := s.repo.AdminBatchReviewWithdrawals(ctx, ids, adminID, input.Status, input.Note); err != nil {
		return nil, err
	}
	return &DistributionBatchWithdrawalReviewResult{Updated: len(ids)}, nil
}
func (s *DistributionService) AdminListWithdrawals(ctx context.Context, filter DistributionAdminWithdrawalListFilter) ([]DistributionWithdrawal, int64, error) {
	return s.repo.AdminListWithdrawals(ctx, normalizeAdminWithdrawalFilter(filter))
}
func (s *DistributionService) AdminGetWithdrawal(ctx context.Context, withdrawalID int64) (*DistributionWithdrawalDetail, error) {
	if withdrawalID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_WITHDRAWAL", "invalid withdrawal")
	}
	return s.repo.AdminGetWithdrawal(ctx, withdrawalID)
}

func (s *DistributionService) AccruePaidOrder(ctx context.Context, input DistributionCommissionInput) (*DistributionCommissionResult, error) {
	if s == nil || s.repo == nil || input.PaymentOrderID <= 0 || input.CustomerUserID <= 0 || !input.ActualPaid.IsPositive() {
		return nil, nil
	}
	input.PaymentCurrency = strings.ToUpper(strings.TrimSpace(input.PaymentCurrency))
	if input.PaymentCurrency == "" {
		input.PaymentCurrency = "CNY"
	}
	if input.PaidAt.IsZero() {
		input.PaidAt = time.Now().UTC()
	}
	return s.repo.AccruePaidOrder(ctx, input)
}

func (s *DistributionService) ReverseRefund(ctx context.Context, paymentOrderID int64, refundedActualAmount decimal.Decimal) error {
	if s == nil || s.repo == nil || paymentOrderID <= 0 || !refundedActualAmount.IsPositive() {
		return nil
	}
	return s.repo.ReverseRefund(ctx, paymentOrderID, refundedActualAmount)
}
