// > go run cr.go -token=$STRAVA_ACCESS_TOKEN -config=config.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/strava/go.strava"
)

type Climb struct {
	Name string `json:"name"`
	Aliases []string `json:"aliases"`
	SegmentId int `json:"segment_id"`
}

type Leaderboards struct {
	MaleOverall []*LeaderboardEntry
	FemaleOverall []*LeaderboardEntry
	MaleYearly []*LeaderboardEntry
	FemaleYearly []*LeaderboardEntry
}

type Data {
	AbsoluteRootUrl string // https://bayarea.climberrankings.com
	CanonicalPath string // /old-la-honda/top-2017-male-riders
	Area string // Bay Area
	Yearly bool
	Male bool
}

type ClimbData {
	Data
	Name string // Old La Honda
	Location string // Woodside, California
	Distance string // 4.88km
	Grade string // 7.91%
	Leaderboard []LeaderboardEntry
}

type LeaderboardEntry {
	Rank int // 2
	RiderId int // 648204
	RiderName string // Alex S
	EffortUrl string // https://www.strava.com/segment_efforts/2920055998
	EffortDate string // 1 Jan 2017
	Score int // 990
}

const MAX_PER_PAGE = 200
const YEAR = time.Now().Year()
const YEARLY = true

var BEGINNING_OF_TIME = time.Unix(0, 0)
var END_OF_TIME = time.Date(9999, time.January, 1, 0, 0, 0, 0, time.UTC)
var START_OF_YEAR = time.Date(YEAR, time.January, 1, 0, 0, 0, 0, time.UTC)
var END_OF_YEAR = time.Date(YEAR, time.December, 31, 23, 59, 59, 999999999, time.UTC)

func main() {
	var accessToken string
	var climbsFilename string
	var climbs []Climb

	// Provide an access token, with write permissions.
	// You'll need to complete the oauth flow to get one.
	flag.StringVar(&accessToken, "token", "", "Access Token")
	flag.StringVar(&climbsFilename, "climbs", "", "Climbs")

	flag.Parse()

	if accessToken == "" {
		fmt.Println("\nPlease provide an access_token, one can be found at https://www.strava.com/settings/api")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if climbs == "" {
		fmt.Println("\nPlease provide a file for which climbs to process.")
		flag.PrintDefaults()
		os.Exit(1)
	}
	raw, err := ioutil.ReadFile(climbsFilename)
	if err != nil {
		fmt.Println("\nNot able to read climbs file: ", climbs, " Error: ", err)
		flag.PrintDefaults()
		os.Exit(1)
	}
	json.Unmarshal(raw, &climbs)

	client := strava.NewClient(accessToken)

	service := strava.NewSegmentsService(client)
	// TODO main index, climb index, about, PERF
	for _, climb := range climbs {
		segment, err := service.Get(climb.SegmentId).Do()
		maybeExit(err)

		slug := slugify(Climb.name)
		// TODO mkdir with parent!

		leaderboards, err := GetLeaderboards(service, climb.SegmentId)
		maybeExit(err)

		maleOverallSlug := slugify(strava.Genders.Male, !YEARLY))
		maybeExit(WriteClimbLeaderboard(climb, segment, leaderboards.MaleOverall, maleOverallSlug)
		maybeExit(WriteClimbLeaderboard(climb, segment, leaderboards.FemaleOverall, slugify(strava.Genders.Female, !YEARLY)))
		maybeExit(WriteClimbLeaderboard(climb, segment, leaderboards.MaleYearly, slugify(strava.Genders.Male, YEARLY)))
		maybeExit(WriteClimbLeaderboard(climb, segment, leaderboards.FemaleYearly, slugify(strava.Genders.Female, YEARLY)))

		fmt.Println("ALIASING '/climbs/", slug, "/index.html' to '/", maleOverallSlug, "/index.html'")
		fmt.Println("ALIASING '/", slug, "/index.html' to '/", maleOverallSlug, "/index.html'")
	}
}


func GetLeaderboards(service *strava.SegmentsService, segmentId int64) (Leaderboards, error) {
	var leaderboards Leaderboards

	// Male Overall
	leaderboard, err := service.ListEfforts(segmentId).
		PerPage(MAX_PER_PAGE).
		Gender(strava.Genders.Male).
		Do()
	if err != nil {
		return nil, err
	}
	leaderboards.MaleOverall = leaderboard

	// Female Overall
	leaderboard, err := service.ListEfforts(segmentId).
		PerPage(MAX_PER_PAGE).
		Gender(strava.Genders.Female).
		Do()
	if err != nil {
		return nil, err
	}
	leaderboards.FemaleOverall = leaderboard

	// Male Yearly
	leaderboard, err := service.ListEfforts(segmentId).
		PerPage(MAX_PER_PAGE).
		Gender(strava.Genders.Male).
		DateRange(strava.DateRanges.ThisYear).
		Do()
	if err != nil {
		return nil, err
	}
	leaderboards.MaleYearly = leaderboard

	leaderboard, err := service.ListEfforts(segmentId).
		PerPage(MAX_PER_PAGE).
		Gender(strava.Genders.Female).
		DateRange(strava.DateRanges.ThisYear).
		Do()
	if err != nil {
		return nil, err
	}
	leaderboards.FemaleYearly = leaderboard

	return leaderboards, nil
}

	//efforts, err := GetAllSortedEfforts(service, segmentId, START_OF_YEAR, END_OF_YEAR)
	//maybeExit(err)
	//athletes, err := GetNAthletesFromEfforts(strava.NewAthletesService(client), MAX_ATHLETES, efforts)
	//maybeExit(err)

	//for i, effort := range efforts {
		//name := fmt.Sprintf("https://www.strava.com/athletes/%v", effort.Athlete.Id)
		//a, _ := athletes[effort.Athlete.Id]
		//if a != nil {
			//name = fmt.Sprintf("%v %v (%s)", a.FirstName, a.LastName, a.Gender)
		//}
		//fmt.Printf("%v) %v - https://www.strava.com/activities/%v = %v\n",
			//i + 1,
			//name,
			//effort.Activity.Id,
			//(time.Duration(effort.ElapsedTime) * time.Second).String())
	//}

func WriteClimbLeaderboard(climb Climb, segment *strava.SegmentDetailed, leaderboard Leaderboard, slug string) error {

	elevationGain := segment.TotalElevationGain
	if elevationGain == 0 {
		elevationGain = segment.ElevationHigh - segment.ElevationLow
	}

	// TODO canonical URL must be an absolute URL

	fmt.Printf("%s: distance: %v gain: %.2f (grade: %.2f%%) mid alt: %.2f\n",
		segment.Name,
		segment.Distance,
		elevationGain,
		(elevationGain / segment.Distance) * 100,
		(segment.ElevationLow + segment.ElevationHigh) / 2)

	fmt.Println("ALIASING '/climbs/", slug, "/index.html' to '/", file, "'")
		for _, alias := slugifyAliases(climb) {
			fmt.Println("ALIASING '/climbs/", alias, "/index.html' to '", file, "'")
			fmt.Println("ALIASING '", alias, "/index.html' to '", file, "'")
		}
}

func slugify(gender strava.Gender, yearly bool) string {
	var path := "top-"
	if yearly {
		path += YEAR + "-"
	}
	if gender == strava.Genders.Male {
		path += "male"
	} else {
		path += "female"
	}
	path += "-riders"
	return slugify(path)
}

func slugifyAliases(climb Climb) []string {
	var slugs []string
	for _, alias := range aliases {
		slugs = append(slugs, slugify(alias)
	}
	return slugs
}

func slugify(str string) string {
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	maybeExit(err)
        slug := reg.ReplaceAllString(str, "-")
        slug = strings.ToLower(strings.Trim(safe, "-"))
	return slug
}

func maybeExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
