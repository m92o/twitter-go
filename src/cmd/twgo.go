//
// twgo.go
//
// twitterパッケージを使用したサンプルプログラム
//
package main

import (
	"twitter"
	"os"
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"
)

func main() {
	var err os.Error
	var statuses []twitter.Status
	var lists []twitter.List
	var acc account
	var user twitter.User
	var users []twitter.User
	var rate twitter.Rate
	var status twitter.Status

	flag.Parse()

	// 設定ファイルから
	if acc, err = readConfFile(); err != nil {
		os.Stderr.WriteString(err.String() + "\n")
        os.Exit(1)
	}
	tw := twitter.NewTwitter(acc.user, acc.password, false)

	switch flag.Arg(0) {
	case "show":
		if status, err = tw.Show(flag.Arg(1)); err == nil {
			showStatus(status)
		}
	case "update":
		err = tw.Update(flag.Arg(1))
	case "destroy":
		err = tw.Destroy(flag.Arg(1))
	case "friends":
		if statuses, err = tw.FriendsTimeline(nil); err == nil {
			showTimeline(statuses)
		}
	case "home":
		if statuses, err = tw.HomeTimeline(nil); err == nil {
			showTimeline(statuses)
		}
	case "user":
		opts := map[string]uint{twitter.OPTION_UserTimeline_Count: 5}
		if statuses, err = tw.UserTimeline(opts); err == nil {
			showTimeline(statuses)
		}
	case "mentions":
		if statuses, err = tw.Mentions(nil); err == nil {
			showTimeline(statuses)
		}
	case "public":
		if statuses, err = tw.PublicTimeline(); err == nil {
			showTimeline(statuses)
		}
	case "lists":
		if lists, err = tw.GetLists(flag.Arg(1), nil); err == nil {
			showLists(lists)
		}
	case "list":
		if statuses, err = tw.ListStatuses(flag.Arg(1), flag.Arg(2), nil); err == nil {
			showTimeline(statuses)
		}
	case "my":
		user, err = tw.VerifyCredentials()
		if err == nil {
			showUser(user)
		}
	case "rate":
		rate, err = tw.RateLimitStatus(false)
		if err == nil {
			showRate(rate)
		}
	case "search":
		if users, err = tw.UsersSearch(flag.Arg(1), nil); err == nil {
			showUsers(users)
		}
	default:
		err = os.ErrorString("Not supported")
	}

	if err != nil {
		os.Stderr.WriteString(err.String() + "\n")
        os.Exit(1)
	}

	os.Exit(0)
}

type account struct {
	user string
	password string
}
func readConfFile() (acc account, err os.Error) {
	var buf []byte

	path := os.Getenv("HOME") + "/.twgo.conf"	// "~/.twgo.conf" だとオープン出来ないので

	if buf, err = ioutil.ReadFile(path); err == nil {
		u, _ := regexp.Compile("USER=\"([^\"]+)\"")
		p, _ := regexp.Compile("PASSWORD=\"([^\"]+)\"")
		if users := u.MatchStrings(string(buf)); users != nil {
			acc.user = users[1]
		}
		if passs := p.MatchStrings(string(buf)); passs != nil {
			acc.password = passs[1]
		}
	}

	return
}

func showTimeline(statuses []twitter.Status) {
	for _, s := range statuses {
		fmt.Println(s.Id, s.Text, s.Created_At, s.Source, s.User.Id)
	}
}

func showUser(u twitter.User) {
	fmt.Printf("Id: %d\n", u.Id)
	fmt.Printf("Name: %s\n", u.Name)
	fmt.Printf("ScreenName: %s\n", u.Screen_Name)
	fmt.Printf("Location: %s\n", u.Location)
	fmt.Printf("Description: %s\n", u.Description)
	fmt.Printf("ProfileImageUrl: %s\n", u.Profile_Image_Url)
	fmt.Printf("Url: %s\n", u.Url)
	fmt.Printf("Protected: %t\n", u.Protected)
	fmt.Printf("FollowersCount: %d\n", u.Followers_Count)
	fmt.Printf("FriendsCount: %d\n", u.Friends_Count)
	fmt.Printf("FavouritesCount: %d\n", u.Favourites_Count)
	fmt.Printf("UtcOffset: %d\n", u.Utc_Offset)
	fmt.Printf("TimeZone: %s\n", u.Time_Zone)
	fmt.Printf("StatusesCount: %d\n", u.Statuses_Count)
}

func showRate(r twitter.Rate) {
	fmt.Printf("RemainingHits: %d\n", r.Remaining_Hits)
	fmt.Printf("HourlyLimit: %d\n", r.Hourly_Limit)
	fmt.Printf("ResetTime: %s\n", r.Reset_Time)
	fmt.Printf("ResetTimeInSeconds: %d\n", r.Reset_Time_In_Seconds)
}

func showStatus(s twitter.Status) {
	fmt.Printf("Id: %d\n", s.Id)
	fmt.Printf("Text: %s\n", s.Text)
	fmt.Printf("CreatedAt: %s\n", s.Created_At)
	fmt.Printf("Source: %s\n", s.Source)
	fmt.Printf("UserId: %d\n", s.User.Id)
}

func showLists(lists []twitter.List) {
	for _, l := range lists {
		fmt.Println(l.Id, l.Name, l.Full_Name, l.Slug, l.Description, l.Member_Count, l.Uri, l.Mode, l.User.Id)
	}
}

func showUsers(users []twitter.User) {
	for _, u := range users {
		fmt.Println(u.Id, u.Name, u.Screen_Name)
	}
}
