package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hebo/mailshine/models"
)

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// NewRedditClient creates a new RedditClient
func NewRedditClient(clientID, clientSecret string) (RedditClient, error) {
	client := RedditClient{clientID: clientID, clientSecret: clientSecret}
	if clientID == "" || clientSecret == "" {
		return client, errors.New("missing client ID or client secret")
	}

	err := client.GetToken()
	return client, err
}

// RedditClient interfaces with reddit
type RedditClient struct {
	accessToken  accessTokenResponse
	clientID     string
	clientSecret string
}

func (r *RedditClient) GetToken() error {
	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	form.Add("device_id", "3v4b553")

	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", bytes.NewBufferString(form.Encode()))
	req.SetBasicAuth(r.clientID, r.clientSecret)
	req.Header.Add("User-agent", "mailshine/v0.1")

	fmt.Println(req.PostForm.Encode())

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	token := accessTokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return err
	}

	fmt.Printf("Token Stored -> %#v\n", token)
	r.accessToken = token
	return nil
}

func (r *RedditClient) FetchSubreddit(subredditName string, numStories int) (listingResponse, error) {
	if r.accessToken.AccessToken == "" {
		panic("no access token")
	}

	listingRes := listingResponse{}
	baseURL, err := url.Parse("https://oauth.reddit.com/")
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		return listingRes, err
	}

	baseURL.Path += fmt.Sprintf("r/%s/top", subredditName)

	// Prepare Query Parameters
	params := url.Values{}
	params.Add("t", "day")
	params.Add("limit", strconv.Itoa(numStories))
	params.Add("raw_json", "1")

	// Add Query Parameters to the URL
	baseURL.RawQuery = params.Encode() // Escape Query Parameters

	fmt.Printf("Encoded URL is %q\n", baseURL.String())

	req, _ := http.NewRequest("GET", baseURL.String(), nil)
	req.Header.Add("User-agent", "mailshine/v0.1")
	req.Header.Add("Authorization", fmt.Sprintf("bearer %s", r.accessToken.AccessToken))

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&listingRes)
	if err != nil {
		return listingRes, err
	}

	// fmt.Printf("List Response-> %+v\n", listingRes)
	log.Printf("Fetched %d stories from %q", len(listingRes.Data.Children), subredditName)
	return listingRes, nil
}

const redditBaseURL = "https://old.reddit.com"

// ToBlock converts to a block
func (l listingResponse) ToBlock(title string) models.Block {
	commentsURL, err := url.Parse(redditBaseURL)
	if err != nil {
		log.Fatal(err)
	}

	block := models.Block{
		Title: title,
	}
	for _, post := range l.Data.Children {
		commentsURL.Path = post.Data.Permalink

		linkURL, err := url.Parse(post.Data.URL)
		if err != nil {
			log.Printf("Failed to parse Link %q: %s", post.Data.URL, err)
		}

		story := models.Story{
			Title:        post.Data.Title,
			Link:         linkURL.String(),
			Hostname:     linkURL.Host,
			CommentsLink: commentsURL.String(),
			NumComments:  post.Data.NumComments,
			Subreddit:    "r/" + post.Data.Subreddit,
		}
		block.Stories = append(block.Stories, story)
	}

	return block
}

type listingResponse struct {
	Kind string `json:"kind"`
	Data struct {
		Modhash  string `json:"modhash"`
		Dist     int    `json:"dist"`
		Children []struct {
			Kind string `json:"kind"`
			Data struct {
				ApprovedAtUtc              interface{}   `json:"approved_at_utc"`
				Subreddit                  string        `json:"subreddit"`
				Selftext                   string        `json:"selftext"`
				AuthorFullname             string        `json:"author_fullname"`
				Saved                      bool          `json:"saved"`
				ModReasonTitle             interface{}   `json:"mod_reason_title"`
				Gilded                     int           `json:"gilded"`
				Clicked                    bool          `json:"clicked"`
				Title                      string        `json:"title"`
				LinkFlairRichtext          []interface{} `json:"link_flair_richtext"`
				SubredditNamePrefixed      string        `json:"subreddit_name_prefixed"`
				Hidden                     bool          `json:"hidden"`
				Pwls                       int           `json:"pwls"`
				LinkFlairCSSClass          interface{}   `json:"link_flair_css_class"`
				Downs                      int           `json:"downs"`
				ThumbnailHeight            int           `json:"thumbnail_height"`
				TopAwardedType             string        `json:"top_awarded_type"`
				HideScore                  bool          `json:"hide_score"`
				Name                       string        `json:"name"`
				Quarantine                 bool          `json:"quarantine"`
				LinkFlairTextColor         string        `json:"link_flair_text_color"`
				UpvoteRatio                float64       `json:"upvote_ratio"`
				AuthorFlairBackgroundColor interface{}   `json:"author_flair_background_color"`
				SubredditType              string        `json:"subreddit_type"`
				Ups                        int           `json:"ups"`
				TotalAwardsReceived        int           `json:"total_awards_received"`
				MediaEmbed                 struct {
				} `json:"media_embed"`
				ThumbnailWidth        int           `json:"thumbnail_width"`
				AuthorFlairTemplateID interface{}   `json:"author_flair_template_id"`
				IsOriginalContent     bool          `json:"is_original_content"`
				UserReports           []interface{} `json:"user_reports"`
				SecureMedia           interface{}   `json:"secure_media"`
				IsRedditMediaDomain   bool          `json:"is_reddit_media_domain"`
				IsMeta                bool          `json:"is_meta"`
				Category              interface{}   `json:"category"`
				SecureMediaEmbed      struct {
				} `json:"secure_media_embed"`
				LinkFlairText       interface{}   `json:"link_flair_text"`
				CanModPost          bool          `json:"can_mod_post"`
				Score               int           `json:"score"`
				ApprovedBy          interface{}   `json:"approved_by"`
				AuthorPremium       bool          `json:"author_premium"`
				Thumbnail           string        `json:"thumbnail"`
				Edited              bool          `json:"-"`
				AuthorFlairCSSClass interface{}   `json:"author_flair_css_class"`
				AuthorFlairRichtext []interface{} `json:"author_flair_richtext"`
				Gildings            struct {
					Gid1 int `json:"gid_1"`
				} `json:"gildings"`
				PostHint            string      `json:"post_hint"`
				ContentCategories   interface{} `json:"content_categories"`
				IsSelf              bool        `json:"is_self"`
				ModNote             interface{} `json:"mod_note"`
				Created             float64     `json:"created"`
				LinkFlairType       string      `json:"link_flair_type"`
				Wls                 int         `json:"wls"`
				RemovedByCategory   interface{} `json:"removed_by_category"`
				BannedBy            interface{} `json:"banned_by"`
				AuthorFlairType     string      `json:"author_flair_type"`
				Domain              string      `json:"domain"`
				AllowLiveComments   bool        `json:"allow_live_comments"`
				SelftextHTML        interface{} `json:"selftext_html"`
				Likes               interface{} `json:"likes"`
				SuggestedSort       interface{} `json:"suggested_sort"`
				BannedAtUtc         interface{} `json:"banned_at_utc"`
				URLOverriddenByDest string      `json:"url_overridden_by_dest"`
				ViewCount           interface{} `json:"view_count"`
				Archived            bool        `json:"archived"`
				NoFollow            bool        `json:"no_follow"`
				IsCrosspostable     bool        `json:"is_crosspostable"`
				Pinned              bool        `json:"pinned"`
				Over18              bool        `json:"over_18"`
				Preview             struct {
					Images []struct {
						Source struct {
							URL    string `json:"url"`
							Width  int    `json:"width"`
							Height int    `json:"height"`
						} `json:"source"`
						Resolutions []struct {
							URL    string `json:"url"`
							Width  int    `json:"width"`
							Height int    `json:"height"`
						} `json:"resolutions"`
						Variants struct {
						} `json:"variants"`
						ID string `json:"id"`
					} `json:"images"`
					Enabled bool `json:"enabled"`
				} `json:"preview"`
				AllAwardings []struct {
					GiverCoinReward          interface{} `json:"giver_coin_reward"`
					SubredditID              interface{} `json:"subreddit_id"`
					IsNew                    bool        `json:"is_new"`
					DaysOfDripExtension      int         `json:"days_of_drip_extension"`
					CoinPrice                int         `json:"coin_price"`
					ID                       string      `json:"id"`
					PennyDonate              interface{} `json:"penny_donate"`
					AwardSubType             string      `json:"award_sub_type"`
					CoinReward               int         `json:"coin_reward"`
					IconURL                  string      `json:"icon_url"`
					DaysOfPremium            int         `json:"days_of_premium"`
					TiersByRequiredAwardings interface{} `json:"tiers_by_required_awardings"`
					ResizedIcons             []struct {
						URL    string `json:"url"`
						Width  int    `json:"width"`
						Height int    `json:"height"`
					} `json:"resized_icons"`
					IconWidth                        int         `json:"icon_width"`
					StaticIconWidth                  int         `json:"static_icon_width"`
					StartDate                        interface{} `json:"start_date"`
					IsEnabled                        bool        `json:"is_enabled"`
					AwardingsRequiredToGrantBenefits interface{} `json:"awardings_required_to_grant_benefits"`
					Description                      string      `json:"description"`
					EndDate                          interface{} `json:"end_date"`
					SubredditCoinReward              int         `json:"subreddit_coin_reward"`
					Count                            int         `json:"count"`
					StaticIconHeight                 int         `json:"static_icon_height"`
					Name                             string      `json:"name"`
					ResizedStaticIcons               []struct {
						URL    string `json:"url"`
						Width  int    `json:"width"`
						Height int    `json:"height"`
					} `json:"resized_static_icons"`
					IconFormat    interface{} `json:"icon_format"`
					IconHeight    int         `json:"icon_height"`
					PennyPrice    interface{} `json:"penny_price"`
					AwardType     string      `json:"award_type"`
					StaticIconURL string      `json:"static_icon_url"`
				} `json:"all_awardings"`
				Awarders                 []interface{} `json:"awarders"`
				MediaOnly                bool          `json:"media_only"`
				CanGild                  bool          `json:"can_gild"`
				Spoiler                  bool          `json:"spoiler"`
				Locked                   bool          `json:"locked"`
				AuthorFlairText          interface{}   `json:"author_flair_text"`
				TreatmentTags            []interface{} `json:"treatment_tags"`
				Visited                  bool          `json:"visited"`
				RemovedBy                interface{}   `json:"removed_by"`
				NumReports               interface{}   `json:"num_reports"`
				Distinguished            interface{}   `json:"distinguished"`
				SubredditID              string        `json:"subreddit_id"`
				ModReasonBy              interface{}   `json:"mod_reason_by"`
				RemovalReason            interface{}   `json:"removal_reason"`
				LinkFlairBackgroundColor string        `json:"link_flair_background_color"`
				ID                       string        `json:"id"`
				IsRobotIndexable         bool          `json:"is_robot_indexable"`
				ReportReasons            interface{}   `json:"report_reasons"`
				Author                   string        `json:"author"`
				DiscussionType           interface{}   `json:"discussion_type"`
				NumComments              int           `json:"num_comments"`
				SendReplies              bool          `json:"send_replies"`
				WhitelistStatus          string        `json:"whitelist_status"`
				ContestMode              bool          `json:"contest_mode"`
				ModReports               []interface{} `json:"mod_reports"`
				AuthorPatreonFlair       bool          `json:"author_patreon_flair"`
				AuthorFlairTextColor     interface{}   `json:"author_flair_text_color"`
				Permalink                string        `json:"permalink"`
				ParentWhitelistStatus    string        `json:"parent_whitelist_status"`
				Stickied                 bool          `json:"stickied"`
				URL                      string        `json:"url"`
				SubredditSubscribers     int           `json:"subreddit_subscribers"`
				CreatedUtc               float64       `json:"created_utc"`
				NumCrossposts            int           `json:"num_crossposts"`
				Media                    interface{}   `json:"media"`
				IsVideo                  bool          `json:"is_video"`
			} `json:"data"`
		} `json:"children"`
		After  string      `json:"after"`
		Before interface{} `json:"before"`
	} `json:"data"`
}
