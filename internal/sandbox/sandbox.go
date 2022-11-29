package sandbox

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/sandbox/ucnuke"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
)

type user struct {
	username string
	password string
	client   *client.Client
}

type Sandbox struct {
	sudo       *client.Client
	user       *user
	httpClient *http.Client
}

// Create creates temporary user and returns client using temporary credentials.
func (s *Sandbox) Create(ctx context.Context) (*client.Client, error) {
	if s.user != nil {
		return s.user.client, nil
	}
	svc := service.New(s.sudo)

	r, err := createSubaccountRequestFromMainAccount(ctx, svc)
	if err != nil {
		return nil, err
	}
	user := user{
		username: r.Subaccount.Username,
		password: r.Subaccount.Password,
	}

	if _, err := svc.CreateSubaccount(ctx, r); err != nil {
		return nil, err
	}
	user.client = client.New(user.username, user.password, client.WithHTTPClient(s.httpClient))
	s.user = &user
	return s.user.client, nil
}

// Delete deletes all resource created by temporary user and removes user.
func (s *Sandbox) Delete(ctx context.Context) error {
	if s.user == nil {
		return nil
	}

	if err := ucnuke.Account(ctx, service.New(s.user.client)); err != nil {
		return err
	}

	return service.New(s.sudo).DeleteSubaccount(ctx, &request.DeleteSubaccountRequest{Username: s.user.username})
}

func New(username, password string) *Sandbox {
	httpClient := http.DefaultClient
	return &Sandbox{sudo: client.New(username, password), user: nil, httpClient: httpClient}
}

func NewWithHTTPClient(username, password string, httpClient *http.Client) *Sandbox {
	return &Sandbox{sudo: client.New(username, password), user: nil, httpClient: httpClient}
}

func createSubaccountRequestFromMainAccount(ctx context.Context, svc *service.Service) (*request.CreateSubaccountRequest, error) {
	account, err := svc.GetAccount(ctx)
	if err != nil {
		return nil, err
	}
	accountDetails, err := svc.GetAccountDetails(ctx, &request.GetAccountDetailsRequest{
		Username: account.UserName,
	})
	if err != nil {
		return nil, err
	}
	if accountDetails.IsSubaccount() {
		return nil, errors.New("only main account can create sandbox environment")
	}
	return &request.CreateSubaccountRequest{
		Subaccount: request.CreateSubaccount{
			Username: tempName(account.UserName),
			Password: password(16),
			Email:    accountDetails.Email,
			Phone:    accountDetails.Phone,
			Country:  accountDetails.Country,
			Currency: accountDetails.Currency,
			Language: accountDetails.Language,
			Timezone: accountDetails.Timezone,
			AllowAPI: 1,
			AllowGUI: 0,
		},
	}, nil
}

// password should be minimum 8 characters with 1 lowercase, 1 uppercase, and 1 number
func password(length int) string {
	chars := [][]rune{
		[]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ"),
		[]rune("abcdefghijklmnopqrstuvwxyz"),
		[]rune("0123456789"),
	}
	p := make([]rune, 0)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length-3; i++ {
		p = append(p, chars[rand.Intn(3)][rand.Intn(len(chars))])
	}
	for i := range chars {
		p = append(p, chars[i][rand.Intn(len(chars[i]))])
	}
	rand.Shuffle(len(p), func(i, j int) {
		p[i], p[j] = p[j], p[i]
	})
	return string(p)
}

func tempName(prefix string) string {
	return prefix + strings.Replace(time.Now().Format("0102150405.000"), ".", "", 1)
}
