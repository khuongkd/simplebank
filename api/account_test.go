package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/khuongkd/simplebank/db/mock"
	db "github.com/khuongkd/simplebank/db/sqlc"
	"github.com/khuongkd/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctr := gomock.NewController(t)
			defer ctr.Finish()

			store := mockdb.NewMockStore(ctr)
			tc.buildStubs(store)
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/account/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateAccount(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		req           db.CreateAcountParams
		buildStubs    func(store *mockdb.MockStore, req db.CreateAcountParams)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			req: db.CreateAcountParams{
				Owner:    account.Owner,
				Currency: account.Currency,
				Balance:  0,
			},
			buildStubs: func(store *mockdb.MockStore, req db.CreateAcountParams) {
				store.EXPECT().
					CreateAcount(gomock.Any(), gomock.Eq(req)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "InternalError",
			req: db.CreateAcountParams{
				Owner:    account.Owner,
				Currency: account.Currency,
				Balance:  0,
			},
			buildStubs: func(store *mockdb.MockStore, req db.CreateAcountParams) {
				store.EXPECT().
					CreateAcount(gomock.Any(), gomock.Eq(req)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			req: db.CreateAcountParams{
				Currency: "xyz",
				Balance:  0,
			},
			buildStubs: func(store *mockdb.MockStore, req db.CreateAcountParams) {
				store.EXPECT().
					CreateAcount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)

			server := NewServer(store)
			tc.buildStubs(store, tc.req)
			recorder := httptest.NewRecorder()
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(tc.req)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/accounts", &buf)
			require.NoError(t, err)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccounts(t *testing.T) {
	listAccount := []db.Account{}
	n := 10
	for i := 0; i < n; i++ {
		account := randomAccount()
		listAccount = append(listAccount, account)
	}

	testCases := []struct {
		name          string
		pageID        int32
		pageSize      int32
		buildStubs    func(store *mockdb.MockStore, params db.ListAccountsParams)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			pageID:   1,
			pageSize: 5,
			buildStubs: func(store *mockdb.MockStore, params db.ListAccountsParams) {
				store.EXPECT().
					ListAccounts(gomock.Any(), params).
					Times(1).
					Return(listAccount[0:5], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchListAccount(t, recorder.Body, listAccount)
			},
		},
		{
			name:     "BadRequest",
			pageID:   0,
			pageSize: 5,
			buildStubs: func(store *mockdb.MockStore, params db.ListAccountsParams) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "NotFound",
			pageID:   1,
			pageSize: 5,
			buildStubs: func(store *mockdb.MockStore, params db.ListAccountsParams) {
				store.EXPECT().
					ListAccounts(gomock.Any(), params).
					Times(1).
					Return([]db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:     "InternalError",
			pageID:   1,
			pageSize: 5,
			buildStubs: func(store *mockdb.MockStore, params db.ListAccountsParams) {
				store.EXPECT().
					ListAccounts(gomock.Any(), params).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		ctrl := gomock.NewController(t)
		store := mockdb.NewMockStore(ctrl)
		server := NewServer(store)
		listAccountsParams := db.ListAccountsParams{
			Limit:  tc.pageSize,
			Offset: (tc.pageID - 1) * tc.pageSize,
		}
		tc.buildStubs(store, listAccountsParams)

		recorder := httptest.NewRecorder()
		request, err := http.NewRequest(http.MethodGet, "/accounts", nil)
		require.NoError(t, err)
		q := request.URL.Query()
		q.Add("page_id", strconv.FormatInt(int64(tc.pageID), 10))
		q.Add("page_size", strconv.FormatInt(int64(tc.pageSize), 10))
		request.URL.RawQuery = q.Encode()

		server.router.ServeHTTP(recorder, request)
		tc.checkResponse(t, recorder)
	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}

func requireBodyMatchListAccount(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotListAccount []db.Account
	err = json.Unmarshal(data, &gotListAccount)
	require.NoError(t, err)
	require.Len(t, gotListAccount, 5)
	for _, account := range gotListAccount {
		require.NotEmpty(t, account)
	}
}
