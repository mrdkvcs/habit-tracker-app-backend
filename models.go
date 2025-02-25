package main

import (
	"github.com/google/uuid"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"time"
)

type ActivityRow interface {
	GetID() uuid.UUID
	GetName() string
	GetPoints() int32
	GetActivityType() string
}

type GetActivitiesRowWrapper struct {
	database.GetActivitiesRow
}

func (g GetActivitiesRowWrapper) GetID() uuid.UUID        { return g.ID }
func (g GetActivitiesRowWrapper) GetName() string         { return g.Name }
func (g GetActivitiesRowWrapper) GetPoints() int32        { return g.Points }
func (g GetActivitiesRowWrapper) GetActivityType() string { return g.ActivityType }

type SetActivityRowWrapper struct {
	database.SetActivityRow
}

func (s SetActivityRowWrapper) GetID() uuid.UUID        { return s.ID }
func (s SetActivityRowWrapper) GetName() string         { return s.Name }
func (s SetActivityRowWrapper) GetPoints() int32        { return s.Points }
func (s SetActivityRowWrapper) GetActivityType() string { return s.ActivityType }

type User struct {
	ID           uuid.UUID   `json:"id"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Username     string      `json:"username"`
	Email        string      `json:"email"`
	PasswordHash interface{} `json:"password_hash"`
	GoogleID     interface{} `json:"google_id"`
}

type SearchedUser struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	HasBeenInvited bool      `json:"has_been_invited"`
}
type SuggestFeature struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Username    string    `json:"username"`
	Upvote      int32     `json:"upvote"`
}

type Activity struct {
	ActivityID uuid.UUID `json:"activity_id"`
	Name       string    `json:"name"`
	Points     int32     `json:"points"`
	Type       string    `json:"type"`
}

type ActivityLog struct {
	ID                  uuid.NullUUID `json:"id"`
	Duration            int32         `json:"duration"`
	Name                interface{}   `json:"name"`
	Points              int32         `json:"points"`
	ActivityDescription string        `json:"activity_description"`
}

type TotalAndAveragePoints struct {
	TotalPoints   string `json:"total_points"`
	AveragePoints string `json:"average_points"`
}

type ProductivityDay struct {
	Date        time.Time   `json:"date"`
	TotalPoints interface{} `json:"total_points"`
	Status      string      `json:"status"`
}

type BestProductivityDay struct {
	Date        time.Time   `json:"date"`
	TotalPoints interface{} `json:"total_points"`
}

type ProductiveUnProductiveTime struct {
	ProductiveTime   int64 `json:"productive_time"`
	UnproductiveTime int64 `json:"unproductive_time"`
}

type ProductivityStats struct {
	ProductivvityPoints        TotalAndAveragePoints      `json:"productivity_points"`
	BestProductivityDay        BestProductivityDay        `json:"best_productivity_day"`
	ProductivityDays           []ProductivityDay          `json:"productivity_days"`
	ProductiveUnProductiveTime ProductiveUnProductiveTime `json:"time"`
}

type DailyProductiveTime struct {
	ProductiveTime   interface{} `json:"productive_time"`
	UnproductiveTime interface{} `json:"unproductive_time"`
}

type RecentActivity struct {
	Duration            int32       `json:"duration"`
	Name                interface{} `json:"name"`
	Points              int32       `json:"points"`
	ActivityDescription string      `json:"activity_description"`
}

type DailyStats struct {
	TotalPoints            interface{}         `json:"total_points"`
	GoalPoints             interface{}         `json:"goal_points"`
	DailyProductiveTime    DailyProductiveTime `json:"daily_productive_time"`
	RecentActivities       []RecentActivity    `json:"recent_activities"`
	DailyActivityLogsCount int64               `json:"daily_activity_logs_count"`
	CurrentStreak          int32               `json:"current_streak"`
	LongestStreak          int32               `json:"longest_streak"`
	StreakMessage          string              `json:"streak_message"`
}

type DailyPoints struct {
	TotalPoints int32 `json:"total_points"`
	GoalPoints  int32 `json:"goal_points"`
}

type Team struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"team_name"`
	TeamIndustry string    `json:"team_industry"`
	TeamSize     int32     `json:"team_size"`
	IsPrivate    bool      `json:"is_private"`
	CreatedBy    uuid.UUID `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TeamActivity struct {
	ActivityName  string   `json:"activity_name"`
	Points        int32    `json:"points"`
	ActivityRoles []string `json:"activity_roles"`
}

type UserTeams struct {
	TeamID   uuid.UUID `json:"id"`
	TeamName string    `json:"team_name"`
	IsOwner  int32     `json:"is_owner"`
}

type TeamInfo struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"team_name"`
	TeamIndustry string    `json:"team_industry"`
	TeamSize     int32     `json:"team_size"`
	IsPrivate    bool      `json:"is_private"`
	CreatedBy    uuid.UUID `json:"created_by"`
}
type Member struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Roles    string    `json:"roles"`
}

type UserTeamActivity struct {
	ID           uuid.UUID `json:"id"`
	TeamID       uuid.UUID `json:"team_id"`
	ActivityName string    `json:"activity_name"`
	Points       int32     `json:"points"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TeamRole struct {
	ID       uuid.UUID `json:"id"`
	RoleName string    `json:"role_name"`
}
type TeamInvitation struct {
	InvitationID uuid.UUID `json:"invitation_id"`
	TeamID       uuid.UUID `json:"team_id"`
	TeamName     string    `json:"team_name"`
	TeamIndustry string    `json:"team_industry"`
	TeamSize     int32     `json:"team_size"`
}

func databaseSuggestFeaturesToSuggestFeatures(dbSuggestFeatures []database.SuggestFeature) []SuggestFeature {
	suggestFeatures := []SuggestFeature{}
	for _, dbSuggestFeature := range dbSuggestFeatures {
		suggestFeature := SuggestFeature{ID: dbSuggestFeature.ID, Title: dbSuggestFeature.Title, Description: dbSuggestFeature.Description, Username: dbSuggestFeature.Username, Upvote: dbSuggestFeature.Upvote}
		suggestFeatures = append(suggestFeatures, suggestFeature)
	}
	return suggestFeatures
}

func databaseActivityLogsToActivityLogs(dbDailyActivityLogs []database.GetDailyActivityLogsRow) []ActivityLog {
	dailyActivityLogs := []ActivityLog{}
	for _, dbDailyActivityLog := range dbDailyActivityLogs {
		if dbDailyActivityLog.Name.Valid {
			dailyActivityLog := ActivityLog{ID: dbDailyActivityLog.ActivityID, Duration: dbDailyActivityLog.Duration, Name: dbDailyActivityLog.Name.String, Points: dbDailyActivityLog.Points, ActivityDescription: dbDailyActivityLog.ActivityDescription}
			dailyActivityLogs = append(dailyActivityLogs, dailyActivityLog)
		} else {
			dailyActivityLog := ActivityLog{ID: dbDailyActivityLog.ActivityID, Duration: dbDailyActivityLog.Duration, Name: nil, Points: dbDailyActivityLog.Points, ActivityDescription: dbDailyActivityLog.ActivityDescription}
			dailyActivityLogs = append(dailyActivityLogs, dailyActivityLog)
		}
	}
	return dailyActivityLogs
}

func databaseTeamInvitationsToTeamInvitations(dbTeamInvitations []database.GetTeamInvitationsRow) []TeamInvitation {
	teamInvitations := []TeamInvitation{}
	for _, dbTeamInvitation := range dbTeamInvitations {
		teamInvitation := TeamInvitation{InvitationID: dbTeamInvitation.InvitationID, TeamID: dbTeamInvitation.TeamID, TeamName: dbTeamInvitation.TeamName, TeamIndustry: dbTeamInvitation.TeamIndustry, TeamSize: dbTeamInvitation.TeamSize}
		teamInvitations = append(teamInvitations, teamInvitation)
	}
	return teamInvitations
}

func databaseMembersToMembers(dbMembers []database.GetTeamMembersRow) []Member {
	members := []Member{}
	for _, dbMember := range dbMembers {
		member := Member{ID: dbMember.ID, Username: dbMember.Username, Roles: dbMember.Roles}
		members = append(members, member)
	}
	return members
}

func databaseUserTeamActivityToUserTeamActivity(dbUserTeamActivities []database.GetUserTeamActivitiesRow) []UserTeamActivity {
	userTeamActivities := []UserTeamActivity{}
	for _, dbuserteamactivity := range dbUserTeamActivities {
		userTeamActivity := UserTeamActivity{ID: dbuserteamactivity.ID, TeamID: dbuserteamactivity.TeamID, ActivityName: dbuserteamactivity.ActivityName, Points: dbuserteamactivity.Points, CreatedAt: dbuserteamactivity.CreatedAt, UpdatedAt: dbuserteamactivity.UpdatedAt}
		userTeamActivities = append(userTeamActivities, userTeamActivity)
	}
	return userTeamActivities
}

func databaseTeamRolesToTeamRoles(dbTeamRoles []database.GetTeamRolesRow) []TeamRole {
	teamroles := []TeamRole{}
	for _, dbteamrole := range dbTeamRoles {
		teamrole := TeamRole{ID: dbteamrole.ID, RoleName: dbteamrole.RoleName}
		teamroles = append(teamroles, teamrole)
	}
	return teamroles
}
func databaseAllTeamRolesToALlTeamRoles(dbTeamRoles []database.GetAllTeamRolesRow) []TeamRole {
	teamroles := []TeamRole{}
	for _, dbteamrole := range dbTeamRoles {
		teamrole := TeamRole{ID: dbteamrole.ID, RoleName: dbteamrole.RoleName}
		teamroles = append(teamroles, teamrole)
	}
	return teamroles
}
func databaseNotAssignedRolesToNotAssignedRoles(dbTeamRoles []database.GetNotAssignedRolesRow) []TeamRole {
	teamroles := []TeamRole{}
	for _, dbteamrole := range dbTeamRoles {
		teamrole := TeamRole{ID: dbteamrole.ID, RoleName: dbteamrole.RoleName}
		teamroles = append(teamroles, teamrole)
	}
	return teamroles
}

func databaseUsersToUsers(dbUsers []database.GetUsersRow) []SearchedUser {
	searchedUsers := []SearchedUser{}
	for _, dbuser := range dbUsers {
		user := SearchedUser{ID: dbuser.ID, Username: dbuser.Username, HasBeenInvited: dbuser.HasBeenInvited}
		searchedUsers = append(searchedUsers, user)
	}
	return searchedUsers
}

func databaseTeamInfoToTeamInfo(dbteaminfo database.GetTeamInFoRow) TeamInfo {
	return TeamInfo{
		ID:           dbteaminfo.ID,
		Name:         dbteaminfo.Name,
		TeamIndustry: dbteaminfo.TeamIndustry,
		TeamSize:     dbteaminfo.TeamSize,
		IsPrivate:    dbteaminfo.IsPrivate,
		CreatedBy:    dbteaminfo.CreatedBy,
	}
}

func databaseUserTeamsToUserTeams(dbuserteams []database.GetUserTeamsRow) []UserTeams {
	userteams := []UserTeams{}
	for _, dbuserteam := range dbuserteams {
		userteam := UserTeams{TeamID: dbuserteam.TeamID, TeamName: dbuserteam.Name, IsOwner: dbuserteam.IsOwner}
		userteams = append(userteams, userteam)
	}
	return userteams
}

func databaseTeamActivityToTeamActivity(dbteamactivities []database.GetTeamActivitiesRow) []TeamActivity {
	teamactivities := []TeamActivity{}
	for _, dbteamactivity := range dbteamactivities {
		teamactivity := TeamActivity{ActivityName: dbteamactivity.ActivityName, Points: dbteamactivity.Points, ActivityRoles: dbteamactivity.ActivityRoles}
		teamactivities = append(teamactivities, teamactivity)
	}
	return teamactivities
}

func databaseTeamToTeam(dbteam database.Team) Team {
	return Team{
		ID:           dbteam.ID,
		Name:         dbteam.Name,
		TeamIndustry: dbteam.TeamIndustry,
		TeamSize:     dbteam.TeamSize,
		IsPrivate:    dbteam.IsPrivate,
		CreatedBy:    dbteam.CreatedBy,
		CreatedAt:    dbteam.CreatedAt,
		UpdatedAt:    dbteam.UpdatedAt,
	}
}

func DatabaseDailyStatsToDailyStats(DBDailyStats DBDailyStats) DailyStats {
	dailyProductiveTime := DailyProductiveTime{
		ProductiveTime:   DBDailyStats.DailyProductiveTime.ProductiveTime,
		UnproductiveTime: DBDailyStats.DailyProductiveTime.UnproductiveTime,
	}
	recentActivities := []RecentActivity{}
	for _, dbRecentActivity := range DBDailyStats.RecentActivities {
		if dbRecentActivity.Name.Valid {
			recentActivity := RecentActivity{Duration: dbRecentActivity.Duration, Name: dbRecentActivity.Name.String, Points: dbRecentActivity.Points, ActivityDescription: dbRecentActivity.ActivityDescription}
			recentActivities = append(recentActivities, recentActivity)
		} else {
			recentActivity := RecentActivity{Duration: dbRecentActivity.Duration, Name: nil, Points: dbRecentActivity.Points, ActivityDescription: dbRecentActivity.ActivityDescription}
			recentActivities = append(recentActivities, recentActivity)
		}
	}
	return DailyStats{
		TotalPoints:            DBDailyStats.TotalPoints,
		GoalPoints:             DBDailyStats.GoalPoints,
		DailyProductiveTime:    dailyProductiveTime,
		RecentActivities:       recentActivities,
		DailyActivityLogsCount: DBDailyStats.DailyActivityLogsCount,
		CurrentStreak:          DBDailyStats.CurrentStreak,
		LongestStreak:          DBDailyStats.LongestStreak,
		StreakMessage:          DBDailyStats.StreakMessage,
	}
}

func databaseDailyPointsToDailyPoints(dbDailyPoints database.GetDailyPointsRow) DailyPoints {
	return DailyPoints{
		TotalPoints: dbDailyPoints.TotalPoints,
		GoalPoints:  dbDailyPoints.GoalPoints,
	}
}

func databaseProductivityStatsToProductivityStats(productivityStats DatabaseProductivityStats) ProductivityStats {
	productivityDays := []ProductivityDay{}
	for _, productivityDay := range productivityStats.ProductivityDays {
		productivityDays = append(productivityDays, ProductivityDay{Date: productivityDay.Date, TotalPoints: productivityDay.TotalPoints, Status: productivityDay.Status})
	}
	totalAveragePoints := TotalAndAveragePoints{
		TotalPoints:   productivityStats.ProductivityPoints.TotalPoints,
		AveragePoints: productivityStats.ProductivityPoints.AveragePointsPerDay,
	}
	bestProductivityDay := BestProductivityDay{
		Date:        productivityStats.BestProductivityDay.Date,
		TotalPoints: productivityStats.BestProductivityDay.TotalPoints,
	}
	productiveUnproductiveTime := ProductiveUnProductiveTime{
		ProductiveTime:   productivityStats.ProductiveUnProductiveTime.ProductiveTime,
		UnproductiveTime: productivityStats.ProductiveUnProductiveTime.UnproductiveTime,
	}
	return ProductivityStats{
		ProductivvityPoints:        totalAveragePoints,
		BestProductivityDay:        bestProductivityDay,
		ProductivityDays:           productivityDays,
		ProductiveUnProductiveTime: productiveUnproductiveTime,
	}
}

func databaseActivitiesToActivities(dbAccs []database.GetActivitiesRow) []Activity {
	activities := []Activity{}
	for _, dbAcc := range dbAccs {
		activity := Activity{ActivityID: dbAcc.ID, Name: dbAcc.Name, Points: dbAcc.Points, Type: dbAcc.ActivityType}
		activities = append(activities, activity)
	}
	return activities
}
func databaseActivityToActivity(dbAcc ActivityRow) Activity {
	return Activity{ActivityID: dbAcc.GetID(), Name: dbAcc.GetName(), Points: dbAcc.GetPoints(), Type: dbAcc.GetActivityType()}
}

func databaseUserToUser(dbuser database.User) User {
	if dbuser.PasswordHash.Valid {
		return User{
			ID:           dbuser.ID,
			CreatedAt:    dbuser.CreatedAt,
			UpdatedAt:    dbuser.UpdatedAt,
			Username:     dbuser.Username,
			Email:        dbuser.Email,
			PasswordHash: dbuser.PasswordHash.String,
			GoogleID:     nil,
		}
	}
	return User{
		ID:           dbuser.ID,
		CreatedAt:    dbuser.CreatedAt,
		UpdatedAt:    dbuser.UpdatedAt,
		Username:     dbuser.Username,
		Email:        dbuser.Email,
		PasswordHash: nil,
		GoogleID:     dbuser.GoogleID.String,
	}
}
