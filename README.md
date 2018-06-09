Climber Rankings compares riders' performances on climbs and maintains
leaderboards ranking performances against each other.

[Strava's API](http://developers.strava.com/docs/reference/) provides the
ability to request the [overall and yearly
leaderboards](https://developers.strava.com/docs/reference/#api-Segments-getLeaderboardBySegmentId).
Upon being started, Climber Rankings reads in a list of climbs and begins
backfilling the segment details and *full* leaderboards for every climb
specified, writing the results into a local database. Climber Rankings takes
care to only query Strava at a rate of 300 req/15min - Strava allows for 600
req/15min, but only 30k requests for a 24 hour period. By querying at 300
req/15min, Climber Rankings can run continuously and not exceed rate limits.
Furthermore, to maintain freshness, the first N*4 requests to Strava are
requests for the first page of the overall and yearly male/female leaderboards
for the N climbs Climber Rankings is configured to rank (except when a new
climb is added, at which point the request for the segment details of the
climb is given priority).

When storing the leaderboard information in the database, Climber Rankings
applies the '[PERF](https://scheibo.github.com/perf)' formula to come up with
a score for each performance. Climber Rankings then builds leaderboards with
the top scorers on each climb broken down into overall/yearly by gender, as
well as leaderboards which rank riders by their average score on M climbs. In
addition, to leaderboards, Climber Rankings creates an index for climbs and
pages for each rider with their respective scores on each of the N climbs
they've rode. There is no index created for riders. Climber Rankings also
supports various 'aliases' for climbs, creating symlinks to ensure the
leaderboards can be accessed by multiple names (though the pages contain a
canonical URL to avoid causing SEO issues).

    / (-> /top-male-riders)
      /top-male-riders
      /top-female-riders
      /2018-top-male-riders
      /2018-top-female-riders
      /top-male-rides
      /top-female-rides
      /2018-top-male-rides
      /2018-top-female-rides

    /climbs
      /climb-a  (/c-a, /climbs/climb-a, /climbs/c-a -> /top-male-riders)
        /top-male-riders
        /top-female-riders
        /2018-top-male-riders
        /2018-top-female-riders

    /riders/12345

Each time Climber Rankings is run it attempts to refresh and backfill its
database, using state stored in the database to determine where it last left
off. If Climber Rankings has a full snapshot of all the leaderboards and
segment details for each climb it will write out the leaderboards as HTML to
be served as static files.

## Usage

    $ git clone https://github.com/scheibo/cr
    $ vim config.json # ...
    $ crontab -l 2>/dev/null; echo "*/15 * * * * go run cr.go -token=$STRAVA_ACCESS_TOKEN" -config=config.json | crontab -

## License

MIT License - Copyright (c) 2018 [Kirk Scheibelhut](https://scheibo.com)
