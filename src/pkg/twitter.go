//
// twitter.go
//
// Twitterクライアントパッケージ
//
// Copyright (c) 2009 Kunio Murasawa <kunio.murasawa@gmail.com>
//
package twitter

import (
	"os";
	"net";
	"http";
	"io";
	"bufio";
	"strconv";
	"strings";
	"fmt";
	"bytes";
	"json";
	"regexp";
	"encoding/base64";
)

// Twitter情報
type Twitter struct {
	Username string;
	Password string;
	UserId string;
	useSsl bool;
}

// ユーザ情報 (全部stringにしちゃったけど良い?)
type User struct {
	Id string;
	Name string;
	ScreenName string;
	Location string;
	Description string;
	ProfileImageUrl string;
	Url string;
	Protected string;
	FollowersCount string;
	FriendsCount string;
	FavouritesCount string;
	UtcOffset string;
	TimeZone string;
	StatusesCount string;
}

// ステータス情報 (全部stringにしちゃったけど良い?)
type Status struct {
	CreatedAt string;
	Id string;
	Text string;
	Source string;
	UserId string;
}

// 標準パッケージのhttpから持ってきた
type readClose struct {
	io.Reader;
	io.Closer;
}

// ユーザ情報一覧 UserIdがキー
var users map[string]User;

// http send
// 標準パッケージのhttp.sendを元に改造した
// Caller should close res.Body when done reading it.
func send(req *http.Request) (res *http.Response, err os.Error) {
	conn, err := net.Dial("tcp", "", req.URL.Host);
	if err != nil {
		return nil, err
	}

	err = req.Write(conn);
	if err != nil {
		conn.Close();
		return nil, err;
	}

	reader := bufio.NewReader(conn);
	res, err = http.ReadResponse(reader);
	if err != nil {
		conn.Close();
		return nil, err;
	}

	r := io.Reader(reader);
	if v := res.GetHeader("Content-Length"); v != "" {
		n, err := strconv.Atoi64(v);
		if err != nil {
			return nil, err;
		}
		r = io.LimitReader(r, n);
	}
	res.Body = readClose{r, conn};

	return;
}

// requestの引数methodに使用
const (
	GET = iota;
	POST;
)
// http request
// Caller should close res.Body when done reading it.
func request(method int, path, body, user, pass string, useSsl bool) (res *http.Response, err os.Error) {
	const HOST = "twitter.com";

	var req http.Request;
	var url string;

	// url
	if useSsl != true {
		url = "http://" + HOST + ":80";
	} else {
		url = "https://" + HOST + ":443";
	}

	if req.URL, err = http.ParseURL(url + path); err != nil {
		return nil, err;
	}

	// method
	switch method {
	case GET:
		req.Method = "GET";
	case POST:
		req.Method = "POST";
	default:
		return nil, os.NewError("invalid method");
	}

	// header
	req.Header = make(map[string]string);

	// authorization
	if user != "" && pass != "" {
		userpass := user + ":" + pass;
		buf := make([]byte, base64.StdEncoding.EncodedLen(len(userpass)));
		base64.StdEncoding.Encode(buf, strings.Bytes(userpass));
		encodedUserpass := string(buf);

		req.Header["Authorization"] = "Basic " + encodedUserpass;
	}

	// content type
	if method == POST {
        req.Header["Content-Type"] = "application/x-www-form-urlencoded";
	}

	// body
	if body != "" {
		req.Body = bytes.NewBufferString(body);
	}

	return send(&req);
}

// http.URLEscapeが直るまでの代わり
func encode(str string) (enc string) {
    var s = "";
    for pos, char := range str {
		switch {
        case char <= 0x007f:
            s = fmt.Sprintf("%c", str[pos]);
		case char >= 0x0080 && char <= 0x07ff:
            b0 := char & 0x07c0 >> 6 + 0xc0;
            b1 := char & 0x003f + 0x80;
            s = fmt.Sprintf("%%%x%%%x", b0, b1);
		case char >= 0x0800 && char <= 0xffff:
            b0 := char & 0xf000 >> 12 + 0xe0;
            b1 := char & 0x0fc0 >> 6 + 0x80;
            b2 := char & 0x003f + 0x80;
            s = fmt.Sprintf("%%%x%%%x%%%x", b0, b1, b2);
		}
        enc += s;
    }
    return enc;
}

func setUser(elem json.Json) {
	var user User;

	user.Id = elem.Get("id").String();
	user.Name = elem.Get("name").String();
	user.ScreenName = elem.Get("screen_name").String();
	user.Location = elem.Get("location").String();
	user.Description = elem.Get("description").String();
	user.ProfileImageUrl = elem.Get("profile_image_url").String();
	user.Url = elem.Get("url").String();
	user.Protected = elem.Get("protected").String();
	user.FollowersCount = elem.Get("followers_count").String();
	user.FriendsCount = elem.Get("friends_count").String();
	user.FavouritesCount = elem.Get("favourites_count").String();
	user.UtcOffset = elem.Get("utc_offset").String();
	user.TimeZone = elem.Get("time_zone").String();
	user.StatusesCount = elem.Get("statuses_count").String();

	users[user.Id] = user;
}

// timeline取得
func timeline(path string, options *map[string]uint, user, pass string, useSsl bool) (statuses []Status, err os.Error) {
	optpath := path;

	// option parameters
	if options != nil {
		for opt, val := range *options {
			optpath += opt + strconv.Uitoa(val);
		}
	}

	res, err := request(GET, optpath, "", user, pass, useSsl);
	if res.StatusCode != 200 {
		err = os.ErrorString(res.Status);
	}

	// response body
	if body, err := io.ReadAll(res.Body); err == nil {
		js, _, _ := json.StringToJson(string(body));
		statuses = make([]Status, js.Len());
		re, _ := regexp.Compile("<a[^>]*>(.*)</a>");
		for i := 0; i < js.Len(); i++ {
			// status
			status := js.Elem(i);
			statuses[i].CreatedAt = status.Get("created_at").String();
			statuses[i].Id = status.Get("id").String();
			statuses[i].Text = status.Get("text").String();
			if srcs := re.MatchStrings(status.Get("source").String()); srcs != nil {
				statuses[i].Source = srcs[1];
			}
			user := status.Get("user");
			id := user.Get("id").String();
			statuses[i].UserId = id;

			// user
			setUser(user);
		}
	}
	res.Body.Close();

	return;
}

// コンストラクタ
func NewTwitter(user, pass string, useSsl bool) *Twitter {
	users = make(map[string]User);

	return &Twitter{user, pass, "", useSsl};
}

// ユーザ情報取得
func (t *Twitter) GetUser(id string) (user *User) {
	var ok bool;

	user = new(User);
	if *user, ok = users[id]; ok == false {
		return nil;
	}

	return;
}

// verify credentials（自分のユーザ情報を取得したい時など）
func (t *Twitter) VerifyCredentials() (err os.Error) {
	const path = "/account/verify_credentials.json";

	res, err := request(GET, path, "", t.Username, t.Password, t.useSsl);
	if res.StatusCode != 200 {
		err = os.ErrorString(res.Status);
	}

	// response body
	if body, err := io.ReadAll(res.Body); err == nil {
		// user
		user, _, _ := json.StringToJson(string(body));
		t.UserId = user.Get("id").String();
		setUser(user);
	}
	res.Body.Close();

	return;
}

// statuses update（つぶやき）
func (t *Twitter) Update(message string) (err os.Error) {
	const (
		path = "/statuses/update.json";
		param = "status=";
	)

// http.URLEscapeはバグってて日本語をエスケープしない
//	body := param + http.URLEscape(message);
	body := param + encode(message);

	res, err := request(POST, path, body, t.Username, t.Password, t.useSsl);
	if res.StatusCode != 200 {
		err = os.ErrorString(res.Status);
	}

	res.Body.Close();

	return;
}

// statuses friends_timeline
func (t *Twitter) PublicTimeline() (statuses []Status, err os.Error) {
	const path = "/statuses/public_timeline.json";

	return timeline(path, nil, "", "", t.useSsl);
}

// FiendsTimelineの引数マップキーに使用
const (
	OPTION_FriendsTimeline_SinceId = "?since_id=";
	OPTION_FriendsTimeline_MaxId = "?max_id=";
	OPTION_FriendsTimeline_Count = "?count=";
	OPTION_FriendsTimeline_Page = "?page=";
)
// statuses friends_timeline
func (t *Twitter) FriendsTimeline(options *map[string]uint) (statuses []Status, err os.Error) {
	const path = "/statuses/friends_timeline.json";

	return timeline(path, options, t.Username, t.Password, t.useSsl);
}

// UserTimelineの引数マップキーに使用
const (
	OPTION_UserTimeline_UserId = "?user_id=";
	OPTION_UserTimeline_ScreenName = "?screen_name=";
	OPTION_UserTimeline_SinceId = "?since_id=";
	OPTION_UserTimeline_MaxId = "?max_id=";
	OPTION_UserTimeline_Count = "?count=";
	OPTION_UserTimeline_Page = "?page=";
)
// statuses user_timeline
func (t *Twitter) UserTimeline(options *map[string]uint) (statuses []Status, err os.Error) {
	const path = "/statuses/user_timeline.json";

	return timeline(path, options, t.Username, t.Password, t.useSsl);
}

// Mentionsの引数マップキーに使用
const (
	OPTION_Mentions_SinceId = "?since_id=";
	OPTION_Mentions_MaxId = "?max_id=";
	OPTION_Mentions_Count = "?count=";
	OPTION_Mentions_Page = "?page=";
)
// statuses mentions
func (t *Twitter) Mentions(options *map[string]uint) (statuses []Status, err os.Error) {
	const path = "/statuses/mentions.json";

	return timeline(path, options, t.Username, t.Password, t.useSsl);
}
