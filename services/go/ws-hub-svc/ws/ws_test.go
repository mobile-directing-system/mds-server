package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

const timeout = 5 * time.Second

type wsRecorder struct {
	received          []json.RawMessage
	toSend            []json.RawMessage
	err               []error
	done              chan struct{}
	closeAfterUpgrade bool
	expectReceived    int
}

func newWSRecorder(toSend []json.RawMessage, expectReceived int) *wsRecorder {
	return &wsRecorder{
		received:       make([]json.RawMessage, 0),
		toSend:         toSend,
		err:            make([]error, 0),
		done:           make(chan struct{}, 0),
		expectReceived: expectReceived,
	}
}

func (wsr *wsRecorder) addError(err error) {
	wsr.err = append(wsr.err, err)
}

func (wsr *wsRecorder) HandlerFunc(ctx context.Context) http.HandlerFunc {
	upgrader := websocket.Upgrader{}
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() { wsr.done <- struct{}{} }()
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			wsr.addError(meh.Wrap(err, "upgrade", nil))
			return
		}
		defer func() { _ = conn.Close() }()
		if wsr.closeAfterUpgrade {
			return
		}
		recorderLifetime, shutdown := context.WithCancel(ctx)
		defer shutdown()
		client := wsutil.NewClient(recorderLifetime, zap.NewNop(), auth.Token{}, conn)
		eg, egCtx := errgroup.WithContext(recorderLifetime)
		var wg sync.WaitGroup
		wg.Add(1)
		eg.Go(func() error {
			defer wg.Done()
			for {
				if len(wsr.received) == wsr.expectReceived {
					return nil
				}
				select {
				case <-egCtx.Done():
					return nil
				case message := <-client.RawConnection().ReceiveRaw():
					wsr.received = append(wsr.received, message)
				}
			}
		})
		wg.Add(1)
		eg.Go(func() error {
			defer wg.Done()
			for _, message := range wsr.toSend {
				select {
				case <-egCtx.Done():
					wsr.addError(meh.NewInternalErr("dropping message to send due to context done", meh.Details{"message": message}))
				case client.RawConnection().SendRaw() <- message:
				}
			}
			return nil
		})
		eg.Go(func() error {
			wg.Wait()
			shutdown()
			<-time.After(500 * time.Millisecond)
			client.Close()
			return nil
		})
		eg.Go(func() error {
			return meh.NilOrWrap(client.RunAndClose(), "run and close", nil)
		})
		err = eg.Wait()
		if err != nil {
			wsr.addError(err)
		}
	}
}

// tokenResolverMock mocks TokenResolver.
type tokenResolverMock struct {
	mock.Mock
}

func (res *tokenResolverMock) ResolvePublicToken(ctx context.Context, publicToken string) (string, error) {
	args := res.Called(ctx, publicToken)
	return args.String(0), args.Error(1)
}

// hubServeSuite tests hub.Serve.
type hubServeSuite struct {
	suite.Suite
}

func (suite *hubServeSuite) TestSingle() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	resolveURL := &url.URL{}
	publicToken := "reach"
	internalToken := "till"
	ginR := testutil.NewGinEngine()
	httpServer := httptest.NewServer(ginR)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	hub := NewNetHub(timeout, zap.NewNop(), resolveURL, map[string]Gate{
		"g": {
			Name: "test_gate",
			Channels: map[wsutil.Channel]Channel{
				"chan_1": {
					URL: fmt.Sprintf("%s/channels/chan_1", serverURL),
				},
			},
		},
	}).(*hub)
	resolver := &tokenResolverMock{}
	hub.tokenResolver = resolver
	resolver.On("ResolvePublicToken", mock.Anything, publicToken).Return(internalToken, nil)
	defer resolver.AssertExpectations(suite.T())
	var wg sync.WaitGroup
	ginR.GET("/ws", func(c *gin.Context) {
		wg.Add(1)
		defer wg.Done()
		headers := make(http.Header)
		err := hub.Serve(c.Writer, c.Request, "g", headers)
		suite.Require().NoError(err, "serve should not fail")
		suite.EqualValues(fmt.Sprintf("Bearer %s", internalToken), headers.Get("Authorization"), "should set correct Authorization-header")
	})
	chan1ToSend := []json.RawMessage{
		json.RawMessage(`{"hello":"world"}`),
		json.RawMessage(`{"i_love":"cookies!"}`),
	}
	chan1 := newWSRecorder(chan1ToSend, 1)
	ginR.GET("/channels/chan_1", func(c *gin.Context) {
		chan1.HandlerFunc(timeout)(c.Writer, c.Request)
	})

	// Connect to the server
	serverConn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/ws", serverURL), nil)
	suite.Require().NoError(err, "should not fail")
	defer func() { _ = serverConn.Close() }()

	// Send auth.
	err = serverConn.WriteMessage(websocket.TextMessage, []byte(publicToken))
	suite.Require().NoError(err, "write client-auth-message should not fail")
	// Send message to server, read response and check to see if it's what we expect.
	err = serverConn.WriteMessage(websocket.TextMessage, []byte(`{"channel":"chan_1","payload":{"for":"my_beautiful_channel"}}`))
	suite.Require().NoError(err, "write client-message should not fail")
	clientReceived := make([]json.RawMessage, 0, 2)
	for range chan1ToSend {
		_, p, err := serverConn.ReadMessage()
		suite.Require().NoError(err, "read client-message should not fail")
		clientReceived = append(clientReceived, p)
	}
	<-chan1.done
	cancel()
	suite.Equal([]json.RawMessage{
		json.RawMessage(`{"channel":"chan_1","payload":{"hello":"world"}}`),
		json.RawMessage(`{"channel":"chan_1","payload":{"i_love":"cookies!"}}`),
	}, clientReceived, "should have received all messages from channel")
	suite.Equal([]json.RawMessage{
		json.RawMessage(`{"for":"my_beautiful_channel"}`),
	}, chan1.received, "channel should have received all messages from client")

	wait()
	wg.Wait()
}

func (suite *hubServeSuite) TestMulti() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	ginR := testutil.NewGinEngine()
	httpServer := httptest.NewServer(ginR)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	hub := NewNetHub(timeout, zap.NewNop(), nil, map[string]Gate{
		"g": {
			Name: "test_gate",
			Channels: map[wsutil.Channel]Channel{
				"chan_1": {
					URL: fmt.Sprintf("%s/channels/chan_1", serverURL),
				},
				"chan_2": {
					URL: fmt.Sprintf("%s/channels/chan_2", serverURL),
				}},
		},
	}).(*hub)
	publicToken := "business"
	resolver := &tokenResolverMock{}
	resolver.On("ResolvePublicToken", mock.Anything, publicToken).Return("", nil)
	defer resolver.AssertExpectations(suite.T())
	hub.tokenResolver = resolver
	ginR.GET("/ws", func(c *gin.Context) {
		err := hub.Serve(c.Writer, c.Request, "g", make(http.Header))
		suite.Require().NoError(err, "serve should not fail")
	})
	chan1 := newWSRecorder([]json.RawMessage{
		json.RawMessage(`{"hello":"world"}`),
		json.RawMessage(`{"i_love":"cookies!"}`),
	}, 2)
	ginR.GET("/channels/chan_1", func(c *gin.Context) {
		chan1.HandlerFunc(timeout)(c.Writer, c.Request)
	})
	chan2 := newWSRecorder([]json.RawMessage{
		json.RawMessage(`{"gustav":"olaf"}`),
	}, 1)
	ginR.GET("/channels/chan_2", func(c *gin.Context) {
		chan2.HandlerFunc(timeout)(c.Writer, c.Request)
	})
	// Connect to the server
	serverConn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/ws", serverURL), nil)
	suite.Require().NoError(err, "should not fail")
	defer func() { _ = serverConn.Close() }()

	// Authenticate.
	err = serverConn.WriteMessage(websocket.TextMessage, []byte(publicToken))
	suite.Require().NoError(err, "write client-auth-message should not fail")
	// Send message to server, read response and check to see if it's what we expect.
	clientReceived := make([]json.RawMessage, 0, 2)
	for i := 0; i < 3; i++ {
		_, p, err := serverConn.ReadMessage()
		suite.Require().NoError(err, "read client-message should not fail")
		clientReceived = append(clientReceived, p)
	}
	err = serverConn.WriteMessage(websocket.TextMessage, []byte(`{"channel":"chan_1","payload":{"for":"my_beautiful_channel"}}`))
	suite.Require().NoError(err, "write client-message should not fail")
	err = serverConn.WriteMessage(websocket.TextMessage, []byte(`{"channel":"chan_2","payload":{"for":"my_beautiful_channel_2"}}`))
	suite.Require().NoError(err, "write client-message should not fail")
	err = serverConn.WriteMessage(websocket.TextMessage, []byte(`{"channel":"chan_1","payload":{"for":"my_beautiful_channel"}}`))
	suite.Require().NoError(err, "write client-message should not fail")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-chan1.done
	}()
	go func() {
		defer wg.Done()
		<-chan2.done
	}()
	wg.Wait()
	cancel()
	suite.Contains(clientReceived, json.RawMessage(`{"channel":"chan_1","payload":{"hello":"world"}}`), "should have received message from channel")
	suite.Contains(clientReceived, json.RawMessage(`{"channel":"chan_1","payload":{"i_love":"cookies!"}}`), "should have received message from channel")
	suite.Contains(clientReceived, json.RawMessage(`{"channel":"chan_2","payload":{"gustav":"olaf"}}`), "should have received message from channel")
	suite.Equal([]json.RawMessage{
		json.RawMessage(`{"for":"my_beautiful_channel"}`),
		json.RawMessage(`{"for":"my_beautiful_channel"}`),
	}, chan1.received, "channel should have received all messages from client")
	suite.Equal([]json.RawMessage{
		json.RawMessage(`{"for":"my_beautiful_channel_2"}`),
	}, chan2.received, "channel should have received all messages from client")
	wait()
}

func (suite *hubServeSuite) TestChannelNotAvailable() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	ginR := testutil.NewGinEngine()
	httpServer := httptest.NewServer(ginR)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	hub := NewNetHub(timeout, zap.NewNop(), nil, map[string]Gate{
		"g": {
			Name: "test_gate",
			Channels: map[wsutil.Channel]Channel{
				"chan_1": {
					URL: fmt.Sprintf("%s/channels/chan_1", serverURL),
				},
				"chan_2": {
					URL: fmt.Sprintf("%s/channels/chan_2", serverURL),
				}},
		},
	}).(*hub)
	resolver := &tokenResolverMock{}
	resolver.On("ResolvePublicToken", mock.Anything, mock.Anything).Return("", nil).Maybe()
	hub.tokenResolver = resolver
	var wg sync.WaitGroup
	ginR.GET("/ws", func(c *gin.Context) {
		wg.Add(1)
		defer wg.Done()
		err := hub.Serve(c.Writer, c.Request, "g", make(http.Header))
		suite.Require().Error(err, "serve should fail")
		c.Status(http.StatusInternalServerError)
	})
	chan1 := newWSRecorder([]json.RawMessage{}, 999)
	ginR.GET("/channels/chan_1", func(c *gin.Context) {
		chan1.HandlerFunc(timeout)(c.Writer, c.Request)
	})
	ginR.GET("/channels/chan_2", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	// Connect to the server
	serverConn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/ws", serverURL), nil)
	suite.Require().NoError(err, "open server conn should not fail")
	defer func() { _ = serverConn.Close() }()
	// Authenticate.
	err = serverConn.WriteMessage(websocket.TextMessage, []byte("court"))
	suite.Require().NoError(err, "write client-auth-message should not fail")

	_, _, err = serverConn.ReadMessage()
	suite.Error(err, "read msesage should fail because of connection being closed by the server")

	cancel()
	wait()
	wg.Wait()
}

func (suite *hubServeSuite) TestChannelUnexpectedClose() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	defer cancel()
	ginR := testutil.NewGinEngine()
	httpServer := httptest.NewServer(ginR)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	hub := NewNetHub(timeout, zap.NewNop(), nil, map[string]Gate{
		"g": {
			Name: "test_gate",
			Channels: map[wsutil.Channel]Channel{
				"chan_1": {
					URL: fmt.Sprintf("%s/channels/chan_1", serverURL),
				},
				"chan_2": {
					URL: fmt.Sprintf("%s/channels/chan_2", serverURL),
				}},
		},
	}).(*hub)
	resolver := &tokenResolverMock{}
	resolver.On("ResolvePublicToken", mock.Anything, mock.Anything).Return("", nil).Maybe()
	hub.tokenResolver = resolver
	var allStuff sync.WaitGroup
	ginR.GET("/ws", func(c *gin.Context) {
		allStuff.Add(1)
		defer allStuff.Done()
		_ = hub.Serve(c.Writer, c.Request, "g", make(http.Header))
	})
	chan1 := newWSRecorder([]json.RawMessage{
		json.RawMessage(`{"hello":"world"}`),
		json.RawMessage(`{"i_love":"cookies!"}`),
		json.RawMessage(`{"i_love":"more_cookies!"}`),
		json.RawMessage(`{"i_love":"and_more_cookies!"}`),
	}, 0)
	ginR.GET("/channels/chan_1", func(c *gin.Context) {
		chan1.HandlerFunc(timeout)(c.Writer, c.Request)
	})
	chan2 := newWSRecorder(nil, 999)
	chan2.closeAfterUpgrade = true
	ginR.GET("/channels/chan_2", func(c *gin.Context) {
		chan2.HandlerFunc(timeout)(c.Writer, c.Request)
	})
	// Connect to the server
	serverConn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/ws", serverURL), nil)
	suite.Require().NoError(err, "should not fail")
	defer func() { _ = serverConn.Close() }()
	// Authenticate.
	err = serverConn.WriteMessage(websocket.TextMessage, []byte("court"))
	suite.Require().NoError(err, "write client-auth-message should not fail")

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-chan1.done
	}()
	go func() {
		defer wg.Done()
		<-chan2.done
	}()
	wg.Wait()
	cancel()
	allStuff.Wait()
	for {
		_, _, err := serverConn.ReadMessage()
		if err == nil {
			continue
		}
		switch err := err.(type) {
		case *websocket.CloseError:
			suite.Equal(websocket.CloseAbnormalClosure, err.Code, "client connection should have been closed")
		case *net.OpError:
			fmt.Printf("%+v", err)
		}
		break
	}
	wait()
}

func Test_hubServe(t *testing.T) {
	suite.Run(t, new(hubServeSuite))
}
