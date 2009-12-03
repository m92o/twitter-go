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
	"io/ioutil";
	"bufio";
	"strconv";
	"strings";
	"fmt";
	"bytes";
	"json";
	"regexp";
	"encoding/base64";
)

// ユーザ情報
type User struct {
	Id uint64;
	Name string;
	Screen_Name string;
	Location string;
	Description string;
	Profile_Image_Url string;
	Url string;
	Protected bool;
	Followers_Count uint;
	Friends_Count uint;
	Favourites_Count uint;
	Utc_Offset int;
	Time_Zone string;
	Statuses_Count uint;
}

// ステータス情報 (全部stringにしちゃったけど良い?)
type Status struct {
	Created_At string;
	Id uint64;
	Text string;
	Source string;
	User User;
}

// List情報 (全部stringにしちゃったけど良い?)
type List struct {
	Id uint64;
	Name string;
	Full_Name string;
	Slug string;
	Description string;
	Member_Count uint;
	Uri string;
	Mode bool;
	User User;
}

// Twitter情報
type Twitter struct {
	Username string;
	Password string;
	useSsl bool;
}
// コンストラクタ
func NewTwitter(user, pass string, useSsl bool) *Twitter {
	return &Twitter{user, pass, useSsl};
}

// verify credentials（自分のユーザ情報を取得したい時など）
func (self *Twitter) VerifyCredentials() (user User, err os.Error) {
	const path = "/account/verify_credentials.json";

	res, err := request(GET, HOST, path, "", self.Username, self.Password, self.useSsl);
	if res.StatusCode != 200 {
		err = os.ErrorString(res.Status);
		return;
	}

	// response body
	if body, err := ioutil.ReadAll(res.Body); err == nil {
		// user
		var u User;
		if ok, errtok := json.Unmarshal(string(body), &u); ok == true {
			user = u;
		} else {
			err = os.ErrorString(errtok);
		}
	}
	res.Body.Close();

	return;
}

type Rate struct {
	Remaining_Hits uint;
	Hourly_Limit uint;
	Reset_Time string;
	Reset_Time_In_Seconds uint;
}
// rate limit status
func (self *Twitter) RateLimitStatus(ipRate bool) (rate Rate, err os.Error) {
	const path = "/account/rate_limit_status.json";
	var user, pass string;

	if ipRate == true {
		// IP's rate
		user = "";
		pass = "";
	} else {
		// User's rate
		user = self.Username;
		pass = self.Password;
	}

	res, err := request(GET, HOST, path, "", user, pass, self.useSsl);
	if res.StatusCode != 200 {
		err = os.ErrorString(res.Status);
		return;
	}

	// response body
	if body, err := ioutil.ReadAll(res.Body); err == nil {
		// rate limit
		var r Rate;
		if ok, errtok := json.Unmarshal(string(body), &r); ok == true {
			rate = r;
		} else {
			err = os.ErrorString(errtok);
		}
	}
	res.Body.Close();

	return;
}

// statuses show
func (self *Twitter) Show(id string) (status Status, err os.Error) {
	path := fmt.Sprintf("/statuses/show/%s.json", id);

	res, err := request(GET, HOST, path, "", self.Username, self.Password, self.useSsl);
	if res.StatusCode != 200 {
		err = os.ErrorString(res.Status);
		return;
	}

	// response body
	if body, err := ioutil.ReadAll(res.Body); err == nil {
		// status
		var s Status;
		if ok, errtok := json.Unmarshal(string(body), &s); ok == true {
			// source
			re, _ := regexp.Compile("<a[^>]*>(.*)</a>");
			if srcs := re.MatchStrings(s.Source); len(srcs) == 2 {
				s.Source = srcs[1];
			}

			status = s;
		} else {
			err = os.ErrorString(errtok);
		}
	}
	res.Body.Close();

	return;
}

// statuses update（つぶやき）
func (self *Twitter) Update(message string) (err os.Error) {
	const (
		path = "/statuses/update.json";
		param = "status=";
	)

// http.URLEscapeはバグってて日本語をエスケープしない
//	body := param + http.URLEscape(message);
	body := param + encode(message);

	res, err := request(POST, HOST, path, body, self.Username, self.Password, self.useSsl);
	if res.StatusCode != 200 {
		err = os.ErrorString(res.Status);
		return;
	}

	res.Body.Close();

	return;
}

// statuses destroy
func (self *Twitter) Destroy(id string) (err os.Error) {
	path := fmt.Sprintf("/statuses/destroy/%s.json", id);

	res, err := request(DELETE, HOST, path, "", self.Username, self.Password, self.useSsl);
	if res.StatusCode != 200 {
		err = os.ErrorString(res.Status);
		return;
	}

	res.Body.Close();

	return;
}

// statuses friends_timeline
func (self *Twitter) PublicTimeline() (statuses []Status, err os.Error) {
	const path = "/statuses/public_timeline.json";

	return self.timeline(path, nil, "", "", self.useSsl);
}

// HomeTimelineの引数マップキーに使用
const (
	OPTION_HomeTimeline_SinceId = "?since_id=";
	OPTION_HomeTimeline_MaxId = "?max_id=";
	OPTION_HomeTimeline_Count = "?count=";
	OPTION_HomeTimeline_Page = "?page=";
)
// statuses home_timeline
func (self *Twitter) HomeTimeline(options map[string] uint) (statuses []Status, err os.Error) {
	const path = "/statuses/home_timeline.json";

	return self.timeline(path, options, self.Username, self.Password, self.useSsl);
}

// FiendsTimelineの引数マップキーに使用
const (
	OPTION_FriendsTimeline_SinceId = "?since_id=";
	OPTION_FriendsTimeline_MaxId = "?max_id=";
	OPTION_FriendsTimeline_Count = "?count=";
	OPTION_FriendsTimeline_Page = "?page=";
)
// statuses friends_timeline
func (self *Twitter) FriendsTimeline(options map[string] uint) (statuses []Status, err os.Error) {
	const path = "/statuses/friends_timeline.json";

	return self.timeline(path, options, self.Username, self.Password, self.useSsl);
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
func (self *Twitter) UserTimeline(options map[string] uint) (statuses []Status, err os.Error) {
	const path = "/statuses/user_timeline.json";

	return self.timeline(path, options, self.Username, self.Password, self.useSsl);
}

// Mentionsの引数マップキーに使用
const (
	OPTION_Mentions_SinceId = "?since_id=";
	OPTION_Mentions_MaxId = "?max_id=";
	OPTION_Mentions_Count = "?count=";
	OPTION_Mentions_Page = "?page=";
)
// statuses mentions
func (self *Twitter) Mentions(options map[string] uint) (statuses []Status, err os.Error) {
	const path = "/statuses/mentions.json";

	return self.timeline(path, options, self.Username, self.Password, self.useSsl);
}

// GetListsの引数マップキーに使用
const (
	OPTION_GetLists_Cursor = "?cursor=";
)
// GET lists
//  user --- UserId or ScreenName
func (self *Twitter) GetLists(user string, options map[string] int) (lists []List, err os.Error) {
	path := fmt.Sprintf("/%s/lists.json", user);

	// option parameters
	if options != nil {
		for opt, val := range options {
			path += opt + strconv.Itoa(val);
		}
	}

	res, err := request(GET, HOST, path, "", self.Username, self.Password, self.useSsl);
	if res.StatusCode != 200 {
		err = os.ErrorString(res.Status);
		return;
	}

	// response body
	if body, err := ioutil.ReadAll(res.Body); err == nil {
		var l []List;
		if ok, errtok := json.Unmarshal(string(body), &l); ok == true {
			lists = l;
		} else {
			err = os.ErrorString(errtok);
		}
	}
	res.Body.Close();

	return;
}

// ListStatusesの引数マップキーに使用
const (
	OPTION_ListStatuses_SinceId = "?since_id=";
	OPTION_ListStatuses_MaxId = "?max_id=";
	OPTION_ListStatuses_PerPage = "?per_page=";
	OPTION_ListStatuses_Page = "?page=";
)
// statuses mentions
//  user --- UserId or ScreenName
//  list --- ListId or ListName
func (self *Twitter) ListStatuses(user, list string, options map[string] uint) (statuses []Status, err os.Error) {
	path := fmt.Sprintf("/%s/lists/%s/statuses.json", user, list);

	return self.timeline(path, options, self.Username, self.Password, self.useSsl);
}

// UsersSearchの引数マップキーに使用
const (
	OPTION_UsersSearch_PerPage = "?per_page=";
	OPTION_UsersSearch_Page = "?page=";
)
// users search
func (self *Twitter) UsersSearch(user string, options map[string] uint) (users []User, err os.Error) {
	path := fmt.Sprintf("/1/users/search.json?q=%s", user);

	// option parameters
	if options != nil {
		for opt, val := range options {
			path += opt + strconv.Uitoa(val);
		}
	}

	res, err := request(GET, HOST, path, "", self.Username, self.Password, self.useSsl);
	if res.StatusCode != 200 {
		err = os.ErrorString(res.Status);
		return;
	}

	// response body
	if body, err := ioutil.ReadAll(res.Body); err == nil {
		var u []User;
		if ok, errtok := json.Unmarshal(string(body), &u); ok == true {
			users = u;
		} else {
			err = os.ErrorString(errtok);
		}
	}
	res.Body.Close();

	return;
}

// 標準パッケージのhttpから持ってきた
type readClose struct {
	io.Reader;
	io.Closer;
}

// http send
// 標準パッケージのhttp.sendを元に改造した（ベーシック認証, SSL等に対応させる為）
// Caller should close res.Body when done reading it.
func send(req *http.Request) (res *http.Response, err os.Error) {
	conn, err := net.Dial("tcp", "", req.URL.Host);
	if err != nil {
		return nil, err;
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
	DELETE;
)
const HOST = "twitter.com";
// http request
// Caller should close res.Body when done reading it.
func request(method int, host, path, body, user, pass string, useSsl bool) (res *http.Response, err os.Error) {
	var req http.Request;
	var url string;

	// url
	if useSsl != true {
		url = "http://" + host + ":80";
	} else {
		url = "https://" + host + ":443";
		return nil, os.NewError("SSL is not implemented yet.");
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
	case DELETE:
		req.Method = "DELETE";
	default:
		return nil, os.NewError("invalid method");
	}

	// header
	req.Header = make(map[string] string);

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

// timeline取得
func (self *Twitter) timeline(path string, options map[string] uint, user, pass string, useSsl bool) (statuses []Status, err os.Error) {
	optpath := path;

	// option parameters
	if options != nil {
		for opt, val := range options {
			optpath += opt + strconv.Uitoa(val);
		}
	}

	res, err := request(GET, HOST, optpath, "", user, pass, useSsl);
	if res.StatusCode != 200 {
		err = os.ErrorString(res.Status);
		return;
	}

	// response body
	if body, err := ioutil.ReadAll(res.Body); err == nil {
		var s []Status;
		if ok, errtok := json.Unmarshal(string(body), &s); ok == true {
			// source
			re, _ := regexp.Compile("<a[^>]*>(.*)</a>");
			for i, sts := range s {
				if srcs := re.MatchStrings(sts.Source); len(srcs) == 2 {
					s[i].Source = srcs[1];
				}
			}

			statuses = s;
		} else {
			err = os.ErrorString(errtok);
		}
	}
	res.Body.Close();

	return;
}
