package templates

type Layout struct {
	CanonicalUrl string
	Title string
	ContentTemplate string
}

type PageNav struct {
	Yearly bool
	AllPath string
	YearlyPath string

	Female bool
	MalePath string
	FemalePath string

	IncludeRides bool
	Rides bool
	RidersPath string
	RidePath string
}

type Climb {
	Title string
	Name string
	Location string
	Distance string
	Grade string
	PageNavTemplate string
	Leaderboard []ClimbLeaderboardEntry
}

type ClimbLeaderboardEntry {
	Rank int
	RiderId int
	RiderName string
	EffortUrl string
	EffortDate string
	EffortDuration string
	Score int
}
