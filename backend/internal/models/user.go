package models

import "time"

// User roles
const (
	RoleClient  = "client"
	RoleAdvisor = "advisor"
)

type User struct {
	ID                 int       `json:"id" db:"id"`
	Email              string    `json:"email" db:"email"`
	Password           string    `json:"-" db:"password_hash"` // Never expose password hash
	Name               string    `json:"name" db:"name"`
	Role               string    `json:"role" db:"role"`
	CreatedByAdvisorID *int      `json:"createdByAdvisorId,omitempty" db:"created_by_advisor_id"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" db:"updated_at"`
}

// IsAdvisor returns true if the user is a financial advisor
func (u *User) IsAdvisor() bool {
	return u.Role == RoleAdvisor
}

// IsClient returns true if the user is a client
func (u *User) IsClient() bool {
	return u.Role == RoleClient
}

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	Role        string `json:"role,omitempty"`        // Optional: "client" or "advisor", defaults to "client"
	InviteToken string `json:"inviteToken,omitempty"` // Optional: for accepting advisor invitation
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type Claims struct {
	UserID int    `json:"userId"`
	Email  string `json:"email"`
}

// AdvisorClient represents the relationship between an advisor and a client
type AdvisorClient struct {
	ID                  int        `json:"id" db:"id"`
	AdvisorID           int        `json:"advisorId" db:"advisor_id"`
	ClientID            int        `json:"clientId" db:"client_id"`
	Status              string     `json:"status" db:"status"` // pending, active, revoked
	AccessLevel         string     `json:"accessLevel" db:"access_level"` // view, edit, full
	InvitationToken     *string    `json:"-" db:"invitation_token"`
	InvitationExpiresAt *time.Time `json:"-" db:"invitation_expires_at"`
	AcceptedAt          *time.Time `json:"acceptedAt,omitempty" db:"accepted_at"`
	CreatedAt           time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time  `json:"updatedAt" db:"updated_at"`
}

// AdvisorClientWithUser includes the client user details
type AdvisorClientWithUser struct {
	AdvisorClient
	Client User `json:"client"`
}

// Access level constants
const (
	AccessLevelView = "view"
	AccessLevelEdit = "edit"
	AccessLevelFull = "full"
)

// Relationship status constants
const (
	RelationshipStatusPending = "pending"
	RelationshipStatusActive  = "active"
	RelationshipStatusRevoked = "revoked"
)

// ClientInvitation represents an invitation sent by an advisor to a client
type ClientInvitation struct {
	ID              int        `json:"id" db:"id"`
	AdvisorID       int        `json:"advisorId" db:"advisor_id"`
	ClientEmail     string     `json:"clientEmail" db:"client_email"`
	InvitationToken string     `json:"-" db:"invitation_token"`
	Status          string     `json:"status" db:"status"` // pending, accepted, expired, cancelled
	ExpiresAt       time.Time  `json:"expiresAt" db:"expires_at"`
	CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
	AcceptedAt      *time.Time `json:"acceptedAt,omitempty" db:"accepted_at"`
}

// ClientInvitationWithAdvisor includes the advisor user details
type ClientInvitationWithAdvisor struct {
	ClientInvitation
	Advisor User `json:"advisor"`
}

// Invitation status constants
const (
	InvitationStatusPending   = "pending"
	InvitationStatusAccepted  = "accepted"
	InvitationStatusExpired   = "expired"
	InvitationStatusCancelled = "cancelled"
)

// SimulationHistory represents a saved Monte Carlo simulation
type SimulationHistory struct {
	ID               int       `json:"id" db:"id"`
	UserID           int       `json:"userId" db:"user_id"`
	RunByUserID      int       `json:"runByUserId" db:"run_by_user_id"`
	Name             *string   `json:"name,omitempty" db:"name"`
	Notes            *string   `json:"notes,omitempty" db:"notes"`
	Params           string    `json:"-" db:"params"`           // JSON stored as string
	Results          string    `json:"-" db:"results"`          // JSON stored as string
	StartingNetWorth float64   `json:"startingNetWorth" db:"starting_net_worth"`
	FinalP50         float64   `json:"finalP50" db:"final_p50"`
	SuccessRate      float64   `json:"successRate" db:"success_rate"`
	TimeHorizonYears int       `json:"timeHorizonYears" db:"time_horizon_years"`
	IsFavorite       bool      `json:"isFavorite" db:"is_favorite"`
	CreatedAt        time.Time `json:"createdAt" db:"created_at"`
}

// SimulationHistoryFull includes parsed params and results
type SimulationHistoryFull struct {
	SimulationHistory
	ParsedParams  *SimulationParams    `json:"params"`
	ParsedResults *MonteCarloResponse  `json:"results"`
	RunByUser     *User                `json:"runByUser,omitempty"`
}

// SimulationHistorySummary is a lightweight version for list views
type SimulationHistorySummary struct {
	ID               int       `json:"id"`
	Name             *string   `json:"name,omitempty"`
	StartingNetWorth float64   `json:"startingNetWorth"`
	FinalP50         float64   `json:"finalP50"`
	SuccessRate      float64   `json:"successRate"`
	TimeHorizonYears int       `json:"timeHorizonYears"`
	IsFavorite       bool      `json:"isFavorite"`
	CreatedAt        time.Time `json:"createdAt"`
	RunByUserName    string    `json:"runByUserName,omitempty"`
}

// ClientNote represents an advisor's note about a client
type ClientNote struct {
	ID        int       `json:"id" db:"id"`
	AdvisorID int       `json:"advisorId" db:"advisor_id"`
	ClientID  int       `json:"clientId" db:"client_id"`
	Note      string    `json:"note" db:"note"`
	Category  string    `json:"category" db:"category"`
	IsPinned  bool      `json:"isPinned" db:"is_pinned"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// ClientNoteWithClient includes the client user details
type ClientNoteWithClient struct {
	ClientNote
	ClientName string `json:"clientName"`
}

// Note category constants
const (
	NoteCategoryGeneral    = "general"
	NoteCategoryMeeting    = "meeting"
	NoteCategoryGoal       = "goal"
	NoteCategoryConcern    = "concern"
	NoteCategoryActionItem = "action_item"
	NoteCategoryPersonal   = "personal"
)

// CreateNoteRequest is the request body for creating a note
type CreateNoteRequest struct {
	ClientID int    `json:"clientId"`
	Note     string `json:"note"`
	Category string `json:"category,omitempty"`
	IsPinned bool   `json:"isPinned,omitempty"`
}

// UpdateNoteRequest is the request body for updating a note
type UpdateNoteRequest struct {
	Note     string `json:"note,omitempty"`
	Category string `json:"category,omitempty"`
	IsPinned *bool  `json:"isPinned,omitempty"`
}

// ClientGoal represents a financial goal set by an advisor for a client (visible to both)
type ClientGoal struct {
	ID            int        `json:"id" db:"id"`
	AdvisorID     int        `json:"advisorId" db:"advisor_id"`
	ClientID      int        `json:"clientId" db:"client_id"`
	Title         string     `json:"title" db:"title"`
	Description   *string    `json:"description,omitempty" db:"description"`
	Category      string     `json:"category" db:"category"`
	Status        string     `json:"status" db:"status"`
	Priority      string     `json:"priority" db:"priority"`
	TargetAmount  *float64   `json:"targetAmount,omitempty" db:"target_amount"`
	CurrentAmount *float64   `json:"currentAmount,omitempty" db:"current_amount"`
	TargetDate    *string    `json:"targetDate,omitempty" db:"target_date"`
	CompletedAt   *time.Time `json:"completedAt,omitempty" db:"completed_at"`
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time  `json:"updatedAt" db:"updated_at"`
}

// Goal category constants
const (
	GoalCategoryRetirement     = "retirement"
	GoalCategorySavings        = "savings"
	GoalCategoryDebt           = "debt"
	GoalCategoryInvestment     = "investment"
	GoalCategoryEducation      = "education"
	GoalCategoryEmergency      = "emergency"
	GoalCategoryMajorPurchase  = "major_purchase"
	GoalCategoryOther          = "other"
)

// Goal status constants
const (
	GoalStatusPending    = "pending"
	GoalStatusInProgress = "in_progress"
	GoalStatusCompleted  = "completed"
	GoalStatusOnHold     = "on_hold"
)

// Goal priority constants
const (
	GoalPriorityLow    = "low"
	GoalPriorityMedium = "medium"
	GoalPriorityHigh   = "high"
)

// CreateGoalRequest is the request body for creating a goal
type CreateGoalRequest struct {
	Title         string   `json:"title"`
	Description   string   `json:"description,omitempty"`
	Category      string   `json:"category,omitempty"`
	Priority      string   `json:"priority,omitempty"`
	TargetAmount  *float64 `json:"targetAmount,omitempty"`
	CurrentAmount *float64 `json:"currentAmount,omitempty"`
	TargetDate    string   `json:"targetDate,omitempty"`
}

// UpdateGoalRequest is the request body for updating a goal
type UpdateGoalRequest struct {
	Title         string   `json:"title,omitempty"`
	Description   *string  `json:"description,omitempty"`
	Category      string   `json:"category,omitempty"`
	Status        string   `json:"status,omitempty"`
	Priority      string   `json:"priority,omitempty"`
	TargetAmount  *float64 `json:"targetAmount,omitempty"`
	CurrentAmount *float64 `json:"currentAmount,omitempty"`
	TargetDate    *string  `json:"targetDate,omitempty"`
}
