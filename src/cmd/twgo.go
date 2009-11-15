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
	"io";
	"regexp";
)

func main() {
	var err os.Error;
	var statuses []twitter.Status;
	var lists []twitter.List;
	var acc account;

	flag.Parse();

	// 設定ファイルから
	if acc, err = readConfFile(); err != nil {
		fmt.Println(err);
        os.Exit(1);
	}
	tw := twitter.NewTwitter(acc.user, acc.password, false);

	switch flag.Arg(0) {
	case "update":
		err = tw.Update(flag.Arg(1));
	case "friends":
		if statuses, err = tw.FriendsTimeline(nil); err == nil {
			showTimeline(statuses);
		}
	case "user":
		opts := map[string]uint{twitter.OPTION_UserTimeline_Count: 5};
		if statuses, err = tw.UserTimeline(&opts); err == nil {
			showTimeline(statuses);
		}
	case "mentions":
		if statuses, err = tw.Mentions(nil); err == nil {
			showTimeline(statuses);
		}
	case "public":
		if statuses, err = tw.PublicTimeline(); err == nil {
			showTimeline(statuses);
		}
	case "lists":
		if lists, err = tw.GetLists(flag.Arg(1), nil); err == nil {
			showLists(lists);
		}
	case "list":
		if statuses, err = tw.ListStatuses(flag.Arg(1), flag.Arg(2), nil); err == nil {
			showTimeline(statuses);
		}
	case "my":
		err = tw.VerifyCredentials();
		if u, ok := tw.Users[tw.UserId]; ok == true {
			showUser(u);
		}
	}

	if err != nil {
		fmt.Println(err);
        os.Exit(1);
	}

	os.Exit(0);
}

type account struct {
	user string;
	password string;
}
func readConfFile() (acc account, err os.Error) {
	var buf []byte;

	// "~/.twgo.conf" だとオープン出来ないので
	path := os.Getenv("HOME") + "/.twgo.conf";

	if buf, err = io.ReadFile(path); err == nil {
		u, _ := regexp.Compile("USER=\"([^\"]+)\"");
		p, _ := regexp.Compile("PASSWORD=\"([^\"]+)\"");
		if users := u.MatchStrings(string(buf)); users != nil {
			acc.user = users[1];
		}
		if passs := p.MatchStrings(string(buf)); passs != nil {
			acc.password = passs[1];
		}
	}

	return;
}

func showTimeline(statuses []twitter.Status) {
	for _, s := range statuses {
		fmt.Println(s.Id, s.Text, s.CreatedAt, s.Source, s.UserId);
	}
}

func showUser(u twitter.User) {
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

func showLists(lists []twitter.List) {
	for _, l := range lists {
		fmt.Println(l.Id, l.Name, l.FullName, l.Slug, l.MemberCount, l.Uri, l.Mode, l.UserId);
	}
}