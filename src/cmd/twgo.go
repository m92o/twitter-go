//
// twgo.go
//
// twitterパッケージを使用したサンプルプログラム
//
package main

import (
	"twitter";
	"os";
	"flag";
	"fmt";
)

//自分のアカウントに書き換えてください
const (
	USER = "";
	PASSWORD = "";
)

func showTimeline(statuses []twitter.Status) {
	for _, s := range statuses {
		fmt.Println(s.Id, s.Text, s.CreatedAt, s.Source, s.UserId);
	}
}

func showUser(u *twitter.User) {
	fmt.Printf("Id: %s\n", u.Id);
	fmt.Printf("Name: %s\n", u.Name);
	fmt.Printf("ScreenName: %s\n", u.ScreenName);
	fmt.Printf("Location: %s\n", u.Location);
	fmt.Printf("Description: %s\n", u.Description);
	fmt.Printf("ProfileImageUrl: %s\n", u.ProfileImageUrl);
	fmt.Printf("Url: %s\n", u.Url);
	fmt.Printf("Protected: %s\n", u.Protected);
	fmt.Printf("FollowersCount: %s\n", u.FollowersCount);
	fmt.Printf("FriendsCount: %s\n", u.FriendsCount);
	fmt.Printf("FavouritesCount: %s\n", u.FavouritesCount);
	fmt.Printf("UtcOffset: %s\n", u.UtcOffset);
	fmt.Printf("TimeZone: %s\n", u.TimeZone);
	fmt.Printf("StatusesCount: %s\n", u.StatusesCount);
}

func main() {
	var err os.Error;
	var statuses []twitter.Status;
	flag.Parse();

	tw := twitter.NewTwitter(USER, PASSWORD, false);

	switch flag.Arg(0) {
	case "update":
		err = tw.Update(flag.Arg(1));
	case "friends":
		opts := map[string]uint{twitter.OPTION_FriendsTimeline_Count: 5};
		if statuses, err = tw.FriendsTimeline(&opts); err == nil {
			showTimeline(statuses);
		}
	case "user":
		opts := map[string]uint{twitter.OPTION_UserTimeline_Count: 5};
		if statuses, err = tw.UserTimeline(&opts); err == nil {
			showTimeline(statuses);
		}
	case "mentions":
		opts := map[string]uint{twitter.OPTION_Mentions_Count: 5};
		if statuses, err = tw.Mentions(&opts); err == nil {
			showTimeline(statuses);
		}
	case "public":
		if statuses, err = tw.PublicTimeline(); err == nil {
			showTimeline(statuses);
		}
	case "my":
		err = tw.VerifyCredentials();
		if u := tw.GetUser(tw.UserId); u != nil {
			showUser(u);
		}
	}

	if err != nil {
		fmt.Println(err);
        os.Exit(1);
	}

	os.Exit(0);
}
