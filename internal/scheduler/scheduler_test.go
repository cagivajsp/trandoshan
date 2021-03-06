package scheduler

import (
	"github.com/creekorful/trandoshan/api"
	"github.com/creekorful/trandoshan/api_mock"
	"github.com/creekorful/trandoshan/internal/messaging"
	"github.com/creekorful/trandoshan/internal/messaging_mock"
	"github.com/golang/mock/gomock"
	"github.com/nats-io/nats.go"
	"testing"
	"time"
)

func TestParseRefreshDelay(t *testing.T) {
	if parseRefreshDelay("") != -1 {
		t.Fail()
	}
	if parseRefreshDelay("50s") != time.Second*50 {
		t.Fail()
	}
	if parseRefreshDelay("50m") != time.Minute*50 {
		t.Fail()
	}
	if parseRefreshDelay("50h") != time.Hour*50 {
		t.Fail()
	}
	if parseRefreshDelay("50d") != time.Hour*24*50 {
		t.Fail()
	}
}

func TestHandleMessageNotOnion(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	apiClientMock := api_mock.NewMockClient(mockCtrl)
	subscriberMock := messaging_mock.NewMockSubscriber(mockCtrl)

	msg := nats.Msg{}
	subscriberMock.EXPECT().
		ReadMsg(&msg, &messaging.URLFoundMsg{}).
		SetArg(1, messaging.URLFoundMsg{URL: "https://example.org"}).
		Return(nil)

	if err := handleMessage(apiClientMock, -1)(subscriberMock, &msg); err == nil {
		t.FailNow()
	}
}

func TestHandleMessageNoSchedule(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	apiClientMock := api_mock.NewMockClient(mockCtrl)
	subscriberMock := messaging_mock.NewMockSubscriber(mockCtrl)

	msg := nats.Msg{}
	subscriberMock.EXPECT().
		ReadMsg(&msg, &messaging.URLFoundMsg{}).
		SetArg(1, messaging.URLFoundMsg{URL: "https://example.onion"}).
		Return(nil)

	apiClientMock.EXPECT().
		SearchResources("https://example.onion", "", time.Time{}, time.Time{}, 1, 1).
		Return([]api.ResourceDto{}, int64(1), nil)

	if err := handleMessage(apiClientMock, -1)(subscriberMock, &msg); err != nil {
		t.FailNow()
	}
}

func TestHandleMessage(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	apiClientMock := api_mock.NewMockClient(mockCtrl)
	subscriberMock := messaging_mock.NewMockSubscriber(mockCtrl)

	msg := nats.Msg{}
	subscriberMock.EXPECT().
		ReadMsg(&msg, &messaging.URLFoundMsg{}).
		SetArg(1, messaging.URLFoundMsg{URL: "https://example.onion"}).
		Return(nil)

	apiClientMock.EXPECT().
		SearchResources("https://example.onion", "", time.Time{}, time.Time{}, 1, 1).
		Return([]api.ResourceDto{}, int64(0), nil)

	subscriberMock.EXPECT().
		PublishMsg(&messaging.URLTodoMsg{URL: "https://example.onion"}).
		Return(nil)

	if err := handleMessage(apiClientMock, -1)(subscriberMock, &msg); err != nil {
		t.FailNow()
	}
}
